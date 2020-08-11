package pbapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/center"
	"github.com/obase/log"
	"github.com/obase/pbapi/access"
	"github.com/obase/pbapi/cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
)

/*
扩展逻辑服务器:
1. 支持proto静态注册
2. 支持api + conf.yml动态注册
3. 支持httpPlugin, httpCache机制

所有逻辑务必确保次序:
- ServiceOptions: server.serviceOptions -> ServiceHandler.serviceOptions -> conf.ServerConfig
- RouterOptions: server.routerOptions -> router.Use()/router.XXX() -> conf.RouterConfig
- ServerOption: conf.ServerPlugins -> server.serverOptions
- Filter: 遵照ServiceOptions规则产生的ServiceSetting的Plugins -> Filter, 也是确保plugins在filter之前, 产生Node.Filter
		遵循RouterOptions产生的RouterSetting的plugins, 其必须在FloatNode.Filter前
*/

func NewServer() *Server {
	server := &Server{
		Router:        newRouter("", nil),
		routerPlugins: make(map[string]RouterPlugin),
		serverPlugins: make(map[string]ServerPlugin),
	}

	// 默认加载的的ServerPlugins
	server.serverPlugins["hostsallow"] = HostsallowServerPlugin
	// 默认加载的FilterPlugins
	server.routerPlugins["hostsallow"] = HostsallowRouterPlugin

	return server
}

type Server struct {
	*Router                                 // 用于兼容gin.IRouter实现, 其中Router.use()设置的filter适用全部入口
	httpRouterCK    []func(*Router)         // http router回调, 兼容旧API, 可以使用router自身API替换
	grpcServerCK    []func(*grpc.Server)    // grpc server回调, 对外扩展
	routerPlugins   map[string]RouterPlugin // 入口过滤器插件机制, 一般仅用于conf.yml的httpRules/plugins设置
	serverPlugins   map[string]ServerPlugin
	routerOptions   []RouterOption
	serverOptions   []grpc.ServerOption
	serviceOptions  []ServiceOption // 默认设置
	serviceHandlers []*ServiceHandler
}

// 重置全部属性,避免占用内存
func (server *Server) dispose() {
	server.Router = nil
	server.httpRouterCK = nil
	server.grpcServerCK = nil
	server.routerPlugins = nil
	server.serverOptions = nil
	server.serviceOptions = nil
	server.serviceHandlers = nil
}

// 用于设置默认的Grpc生成规则, 目前会自动开启所有method的grpc, http. 但是wbsk需要配置
func defaultServiceSetting(sh *ServiceHandler) *ServiceSetting {
	s := &ServiceSetting{
		PackageName: sh.PackageName,
		ServiceName: sh.ServiceName,
		Methods:     make(map[string]*MethodSetting),
	}
	s.GrpcOff = false
	for k, _ := range sh.Adapters {
		path := DefaultPathGenerator(sh.PackageName, sh.ServiceName, k)
		s.Methods[k] = &MethodSetting{
			HttpOff:  false,
			HttpPath: path,
			WbskOff:  true,
			WbskPath: path,
		}
	}
	return s
}

// 用于生成默认的router
const (
	ServiceSuffix       = "Service"
	ServiceSuffixLength = len(ServiceSuffix)
)

func defaultRouterSetting(packageName, serviceName, methodName string, path string, method string) *RouterSetting {
	if path == "" {
		path = DefaultPathGenerator(packageName, serviceName, methodName)
	}
	return &RouterSetting{
		PackageName: packageName,
		ServiceName: serviceName,
		MethodName:  methodName,
		Path:        path,
		Method:      method,
	}
}

/* 补充gin的IRouter路由信息*/
func (server *Server) WithRouter(rf ...func(*Router)) {
	server.httpRouterCK = append(server.httpRouterCK, rf...)
}

func (server *Server) WithServer(rf ...func(server *grpc.Server)) {
	server.grpcServerCK = append(server.grpcServerCK, rf...)
}

func (server *Server) ServerOption(options ...grpc.ServerOption) {
	server.serverOptions = append(server.serverOptions, options...)
}

func (server *Server) ServiceOption(options ...ServiceOption) {
	server.serviceOptions = append(server.serviceOptions, options...)
}

func (server *Server) RouterOption(options ...RouterOption) {
	server.routerOptions = append(server.routerOptions, options...)
}

func (server *Server) RouterPlugin(name string, rf RouterPlugin) {
	server.routerPlugins[name] = rf
}

func (server *Server) ServerPlugin(name string, rf ServerPlugin) {
	server.serverPlugins[name] = rf
}

func (server *Server) RegisterService(handler RegisterServiceHandler, service interface{}, options ...ServiceOption) {
	sdesc, pname, sname, adapters := handler(service)
	server.serviceHandlers = append(server.serviceHandlers, &ServiceHandler{
		ServiceDesc: sdesc,
		ServiceImpl: service,
		PackageName: pname,
		ServiceName: sname,
		Adapters:    adapters,
		Options:     options,
	})
}

func (server *Server) Serve() error {
	return server.ServeWith(LoadConfig())
}

func (server *Server) ServeWith(config *Config) error {

	config = mergeConfig(config)

	// 没有配置任何启动,直接退出. 注意: 没有默认80之类的设置
	if config.GrpcPort == 0 && config.HttpPort == 0 {
		return nil
	}

	var (
		grpcServer   *grpc.Server
		grpcListener net.Listener
		httpServer   *http.Server
		httpListener net.Listener
		httpCache    cache.Cache
		cancelfun    context.CancelFunc
		accesslog    *log.Logger
		err          error
		grpcfunc     func() // 用于延迟启动
		httpfunc     func() // 用于延迟启动

	)

	runctx, cancelfun := context.WithCancel(context.Background())
	defer func() {
		log.Flush()
		// 反注册consul服务,另外还设定了超时反注册,双重保障
		if config.Name != "" {
			deregisterService(config)
		}
		// 退出需要明确关闭
		if grpcListener != nil {
			grpcListener.Close()
		}
		if httpListener != nil {
			httpListener.Close()
		}
		if httpCache != nil {
			httpCache.Close()
		}
		if cancelfun != nil {
			cancelfun()
		}
		if accesslog != nil {
			accesslog.Close()
		}
	}()

	// 计算setting
	for _, handler := range server.serviceHandlers {
		// 生成默认
		ss := defaultServiceSetting(handler)
		// merge全局
		for _, so := range server.serviceOptions {
			so(ss)
		}
		// merge局部
		for _, so := range handler.Options {
			so(ss)
		}
		// merge配置
		for _, so := range MergeServerConfig(config.ServerConfig) {
			so(ss)
		}
		handler.setting = ss
	}

	// 创建grpc服务器
	if config.GrpcPort > 0 {

		var serverOptions []grpc.ServerOption
		if len(config.ServerPlugins) > 0 {
			for _, v := range config.ServerPlugins {
				if len(v) > 0 {
					if plugin := server.serverPlugins[v[0]]; plugin != nil {
						if option := plugin(v[1:]); option != nil {
							serverOptions = append(serverOptions, option)
						}
					} else {
						return errors.New(fmt.Sprintf("invalid server plugin: %v", v))
					}
				}
			}
		}
		if len(server.serverOptions) > 0 {
			for _, option := range server.serverOptions {
				if option != nil {
					serverOptions = append(serverOptions, option)
				}
			}
		}
		// 设置keepalive超时
		if config.GrpcKeepAlive != 0 {
			serverOptions = append(serverOptions, grpc.KeepaliveParams(keepalive.ServerParameters{
				Time: config.GrpcKeepAlive,
			}))
		}
		grpcServer = grpc.NewServer(serverOptions...)

		// 安装grpc相关配置
		for _, handler := range server.serviceHandlers {
			if !handler.setting.GrpcOff {
				grpcServer.RegisterService(handler.ServiceDesc, handler.ServiceImpl)
			}
		}
		for _, ck := range server.grpcServerCK {
			ck(grpcServer) // 附加额外的Grpc设置,预防额外逻辑
		}
		// 注册grpc服务
		if config.Name != "" {
			registerServiceGrpc(grpcServer, config)
		}
		// 创建监听端口
		grpcListener, err = graceListenGrpc(config.GrpcHost, config.GrpcPort)
		if err != nil {
			log.Errorf("grpc server listen error: %v", err)
			log.Flush()
			return err
		}
		// 启动grpc服务
		grpcfunc = func() {
			if err = grpcServer.Serve(grpcListener); err != nil {
				log.Errorf("grpc server serve error: %v", err)
				log.Flush()
				os.Exit(1)
			}
		}
	}

	// 创建http服务器
	if config.HttpPort > 0 {

		// 安装http相关配置
		var upgrader *websocket.Upgrader
		for _, handler := range server.serviceHandlers {
			for mname, adapt := range handler.Adapters {
				ms := handler.setting.Methods[mname]
				// for http
				if ms == nil || !ms.HttpOff {
					// 确保plugins优先filter
					var filter gin.HandlersChain
					if len(ms.HttpPlugins) > 0 {
						for _, v := range ms.HttpPlugins {
							if len(v) > 0 {
								plugin := server.routerPlugins[v[0]]
								if plugin != nil {
									for _, f := range plugin(v[1:]) {
										if f != nil {
											filter = append(filter, f)
										}
									}
								} else {
									return errors.New(fmt.Sprintf("invalid router plugin: %v", v))
								}
							}
						}
					}
					if len(ms.HttpFilter) > 0 {
						filter = append(filter, ms.HttpFilter...)
					}
					server.Router.handle(MethodPost, ms.HttpPath, &Node{
						PackageName: handler.PackageName,
						ServiceName: handler.ServiceName,
						MethodName:  mname,
						Filter:      filter,
						Handler:     CreateHandlerFunc4Http(handler.ServiceName+"."+mname, adapt),
					})
				}
				// for wbsk
				if ms == nil || !ms.WbskOff {
					// http get实现websocket
					if upgrader == nil {
						upgrader = CreateWebsocketUpgrader(config)
					}
					// 确保plugins优先filter
					var filter gin.HandlersChain
					if len(ms.WbskPlugins) > 0 {
						for _, v := range ms.WbskPlugins {
							if len(v) > 0 {
								plugin := server.routerPlugins[v[0]]
								if plugin != nil {
									for _, f := range plugin(v[1:]) {
										if f != nil {
											filter = append(filter, f)
										}
									}
								} else {
									return errors.New(fmt.Sprintf("invalid router plugin: %v", v))
								}
							}
						}
					}
					if len(ms.WbskFilter) > 0 {
						filter = append(filter, ms.WbskFilter...)
					}

					server.Router.handle(MethodGet, ms.WbskPath, &Node{
						PackageName: handler.PackageName,
						ServiceName: handler.ServiceName,
						MethodName:  mname,
						Filter:      filter,
						Handler:     CreateHandlerFunc4Wbsk(handler.ServiceName+"."+mname, upgrader, adapt),
					})
				}
			}
		}

		for _, ck := range server.httpRouterCK {
			ck(server.Router)
		}
		httpCache = cache.New(config.Cache)
		accesslog, err = access.NewLogger(runctx, config.Accesslog)
		if err != nil {
			log.Errorf("create access logger error: %v", err)
			return err
		}
		// 核心转换生成ServeMux
		engine, err := server.compileRouterEngine(config, httpCache, access.NewHandlerFunc(accesslog))
		if err != nil {
			log.Errorf("http server compile error: %v", err)
			return err
		}
		// 最后才注册,避免前面的安全机制影响
		if config.Name != "" {
			registerServiceHttp(engine, config)
		}

		httpServer = &http.Server{
			Handler: engine,
		}
		// 创建监听端口
		httpListener, err = graceListenHttp(config.HttpHost, config.HttpPort, config.HttpKeepAlive)
		if err != nil {
			log.Errorf("http server listen error: %v", err)
			return err
		}
		// 支持TLS,或http2.0
		if config.HttpCertFile != "" {
			httpfunc = func() {
				if err := httpServer.ServeTLS(httpListener, config.HttpCertFile, config.HttpKeyFile); err != nil {
					log.Errorf("http server serve error: %v", err)
					log.Flush()
					os.Exit(1)
				}
			}
		} else {
			httpfunc = func() {
				if err := httpServer.Serve(httpListener); err != nil {
					log.Errorf("http server serve error: %v", err)
					log.Flush()
					os.Exit(1)
				}
			}
		}
	}
	// 释放无用缓存
	server.dispose()

	// 延迟启动
	if grpcfunc != nil {
		go grpcfunc()
	}
	if httpfunc != nil {
		go httpfunc()
	}
	// 优雅关闭http与grpc服务
	graceShutdownOrRestart(grpcServer, grpcListener, httpServer, httpListener)

	return nil
}

func (server *Server) compileRouterEngine(config *Config, cache cache.Cache, accesslog gin.HandlerFunc) (*gin.Engine, error) {

	// 确保conf.yml的routerConfig可以覆盖server.routerOptions
	var routerOptions = MergeRouterConfig(config.RouterConfig)

	// 遍历所有结点,处理plugins/cache/access/off, 但不包括proxy. 因为它涉及替换与新加
	var flatnodes = server.Router.Flattern()

	// 第1步处理非proxy的结点
	for _, node := range flatnodes {
		// 创建默认的RouterSetting
		rs := defaultRouterSetting(node.PackageName, node.ServiceName, node.MethodName, node.Path, node.Method)
		// 确保conf.yml的routerConfig在server.routerOptions后执行
		for _, option := range server.routerOptions {
			if option != nil {
				option(rs)
			}
		}
		for _, option := range routerOptions {
			if option != nil {
				option(rs)
			}
		}

		if rs.Off || rs.ProxyPath != "" {
			node.Off = true // 如果是关闭或被代理的结点都不再启动
		} else {
			node.Plugins = rs.Plugins
			node.Cache = rs.Cache
			node.Access = rs.Access
		}
	}

	// 第2步附加需proxy的结点
	for _, rc := range config.RouterConfig {
		// 设置了代理并且没有关闭
		if rc.ProxyPath != "" && !rc.Off {
			// 必须先剔除已经禁用的方法
			for _, method := range rc.Methods {
				var node *FlatNode
				if rc.ProxyService == "" {
					node = cloneProxyTarget(flatnodes, rc.ProxyPath, method)
					if node == nil {
						return nil, errors.New(fmt.Sprintf("invalid proxy target %v", rc.ProxyPath))
					}
					// 修订为被代理路径
					node.Path = rc.Path
					node.Method = method
				} else {
					// 外部代理
					var proxy *httputil.ReverseProxy
					if rc.ProxyHttps {
						proxy = center.HttpProxyHandler(rc.ProxyService, rc.ProxyPath)
					} else {
						proxy = center.HttpProxyHandler(rc.ProxyService, rc.ProxyPath)
					}

					node = &FlatNode{
						Path:   rc.Path,
						Method: method,
						Handler: func(ctx *gin.Context) {
							proxy.ServeHTTP(ctx.Writer, ctx.Request)
						},
					}
				}
				node.Plugins = rc.Plugins
				node.Cache = rc.Cache
				node.Access = rc.Access

				flatnodes = append(flatnodes, node)
			}
		}
	}

	// 至此,floatnode包含了所有结点(包括off),对称转换为engine的相关操作
	gin.SetMode(gin.ReleaseMode) // 线上应该设置为release模式
	engine := gin.New()

	var routerFilter gin.HandlersChain
	for _, v := range config.RouterPlugins {
		if len(v) > 0 {
			plugin := server.routerPlugins[v[0]]
			if plugin != nil {
				for _, f := range plugin(v[1:]) {
					if f != nil {
						routerFilter = append(routerFilter, f)
					}
				}
			} else {
				return nil, errors.New(fmt.Sprintf("invalid router plugin: %v", v))
			}
		}
	}

	for _, node := range flatnodes {
		if node.Off {
			continue
		}
		/*相关filter次序: config.routerPlugins > flatnode.plugins(来自routerConfig) > flatnode.filter(来自Service或者Router)*/
		var nodeFilter gin.HandlersChain
		for _, v := range node.Plugins {
			if len(v) > 0 {
				plugin := server.routerPlugins[v[0]]
				if plugin != nil {
					for _, f := range plugin(v[1:]) {
						if f != nil {
							nodeFilter = append(nodeFilter, f)
						}
					}
				} else {
					return nil, errors.New(fmt.Sprintf("invalid router plugin: %v", v))
				}
			}
		}

		switch node.Method {
		case MethodStatic:
			if len(routerFilter) > 0 || len(nodeFilter) > 0 || len(node.Filter) > 0 {
				// 间接实现filter
				gpath, spath := SplitBase(node.Path)
				group := engine.Group(gpath, newHandlersChain(node.Access, accesslog, routerFilter, nodeFilter, node.Filter, nil)...)
				group.Static(spath, node.File)
			} else {
				engine.Static(node.Path, node.File)
			}
		case MethodStaticFile:
			if len(routerFilter) > 0 || len(nodeFilter) > 0 || len(node.Filter) > 0 {
				// 间接实现filter
				gpath, spath := SplitBase(node.Path)
				group := engine.Group(gpath, newHandlersChain(node.Access, accesslog, routerFilter, nodeFilter, node.Filter, nil)...)
				group.Static(spath, node.File)
			} else {
				engine.StaticFile(node.Path, node.File)
			}
		case MethodStaticFS:
			if len(routerFilter) > 0 || len(nodeFilter) > 0 || len(node.Filter) > 0 {
				// 间接实现filter
				gpath, spath := SplitBase(node.Path)
				group := engine.Group(gpath, newHandlersChain(node.Access, accesslog, routerFilter, nodeFilter, node.Filter, nil)...)
				group.Static(spath, node.File)
			} else {
				engine.StaticFS(node.Path, node.FileSystem)
			}
		default:
			engine.Handle(node.Method, node.Path, newHandlersChain(node.Access, accesslog, routerFilter, nodeFilter, node.Filter, node.Handler)...)
		}
	}
	return engine, nil
}

func cloneProxyTarget(fnodes []*FlatNode, path string, method string) *FlatNode {
	for _, fnode := range fnodes {
		if path == fnode.Path && method == fnode.Method {
			var result = *fnode
			return &result
		}
	}
	return nil
}

func newHandlersChain(access bool, accesslog gin.HandlerFunc, c1 gin.HandlersChain, c2 gin.HandlersChain, c3 gin.HandlersChain, h gin.HandlerFunc) gin.HandlersChain {
	var ret gin.HandlersChain
	if access && accesslog != nil {
		ret = append(ret, accesslog)
	}
	if len(c1) > 0 {
		ret = append(ret, c1...)
	}
	if len(c2) > 0 {
		ret = append(ret, c2...)
	}
	if len(c3) > 0 {
		ret = append(ret, c3...)
	}
	if h != nil {
		ret = append(ret, h)
	}
	return ret
}
