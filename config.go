package pbapi

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/obase/conf"
	"github.com/obase/pbapi/access"
	"github.com/obase/pbapi/cache"
	"google.golang.org/grpc"
	"time"
)

type RouterConfig struct {
	Package       string     `json:"package" bson:"package" yaml:"package"`
	Service       string     `json:"service" bson:"service" yaml:"service"`
	Method        string     `json:"method" bson:"method" yaml:"method"`                      // service method
	Path          string     `json:"path" bson:"path" yaml:"path"`                            //路径模式
	Methods       []string   `json:"methods" bson:"methods" yaml:"methods"`                   // 请求方法http method
	ProxyPath     string     `json:"proxyPath" bson:"proxyPath" yaml:"proxyPath"`             //代理路径
	ProxyService  string     `json:"proxyService" bson:"proxyService" yaml:"proxyService"`    // 代理外部服务
	ProxyHttps    bool       `json:"proxyHttps" bson:"proxyHttps" yaml:"proxyHttps"`          // 代理使用https
	Plugins       [][]string `json:"plugins" bson:"plugins" yaml:"plugins"`                   // expression: name(param1,param2,...)
	Cache         int64      `json:"cache" bson:"cache" yaml:"cache"`                         // 缓存秒数
	Off           bool       `json:"off" bson:"off" yaml:"off"`                               // 临时禁用
	Access        bool       `json:"access" bson:"access" yaml:"access"`                      // 是否打印access log
	SetProxyHttps bool       `json:"setProxyHttps" bson:"setProxyHttps" yaml:"setProxyHttps"` // 是否设置过ProxyHttps
	SetOff        bool       `json:"setOff" bson:"setOff" yaml:"setOff"`                      // 是否设置过Off, 否则只有true才设置
	SetAccess     bool       `json:"setAccess" bson:"setAccess" yaml:"setAccess"`             // 是否设置了Access,否则只有true才设置
}

type ServerConfig struct {
	Package     string     `json:"package" bson:"package" yaml:"package"`
	Service     string     `json:"service" bson:"service" yaml:"service"`
	Method      string     `json:"method" bson:"method" yaml:"method"`
	GrpcOff     bool       `json:"grpcOff" bson:"grpcOff" yaml:"grpcOff"`
	HttpOff     bool       `json:"httpOff" bson:"httpOff" yaml:"httpOff"`
	HttpPath    string     `json:"httpPath" bson:"httpPath" yaml:"httpPath"` // ServerPathDefault(packageName, serviceName, methodName) string
	HttpPlugins [][]string `json:"httpPlugins" bson:"httpPlugins" yaml:"httpPlugins"`
	WbskOff     bool       `json:"wbskOff" bson:"wbskOff" yaml:"wbskOff"`
	WbskPath    string     `json:"wbskPath" bson:"wbskPath" yaml:"wbskPath"` // ServerPathDefault(packageName, serviceName, methodName)
	WbskPlugins [][]string `json:"wbskPlugins" bson:"wbskPlugins" yaml:"wbskPlugins"`
	SetGrpcOff  bool       `json:"setGrpcOff" bson:"setGrpcOff" yaml:"setGrpcOff"` // 是否设置了GrpcOff, 否则只有true才设置
	SetHttpOff  bool       `json:"setHttpOff" bson:"setHttpOff" yaml:"setHttpOff"` // 是否设置了HttpOff, 否则只有true才设置
	SetWbskOff  bool       `json:"setWbskOff" bson:"setWbskOff" yaml:"setWbskOff"` // 是否设置了WbskOff, 否则只有true才设置
}

/*服务配置,注意兼容性.Grpc服务添加前缀"grpc."*/
type Config struct {
	Name                string            `json:"name" bson:"name" yaml:"name"`                                              // 注册服务名,如果没有则不注册
	HttpHost            string            `json:"httpHost" bson:"httpHost" yaml:"httpHost"`                                  // Http暴露主机,默认首个私有IP
	HttpPort            int               `json:"httpPort" bson:"httpPort" yaml:"httpPort"`                                  // Http暴露端口, 默认80
	HttpKeepAlive       time.Duration     `json:"httpKeepAlive" bson:"httpKeepAlive" yaml:"httpKeepAlive"`                   // Keepalive
	HttpCheckTimeout    string            `json:"httpCheckTimeout" bson:"httpCheckTimeout" yaml:"httpCheckTimeout"`          // 注册服务心跳检测超时
	HttpCheckInterval   string            `json:"httpCheckInterval" bson:"httpCheckInterval" yaml:"httpCheckInterval"`       // 注册服务心跳检测间隔
	HttpCertFile        string            `json:"httpCertFile" bson:"httpCertFile" yaml:"httpCertFile"`                      // 启用TLS
	HttpKeyFile         string            `json:"httpKeyFile" bson:"httpKeyFile" yaml:"httpKeyFile"`                         // 启用TLS
	WbskReadBufferSize  int               `json:"wbskReadBufferSize" bson:"wbskReadBufferSize" yaml:"wbskReadBufferSize"`    // 默认4092
	WbskWriteBufferSize int               `json:"wbskWriteBufferSize" bson:"wbskWriteBufferSize" yaml:"wbskWriteBufferSize"` // 默认4092
	WbskNotCheckOrigin  bool              `json:"wbskNotCheckOrigin" bson:"wbskNotCheckOrigin" yaml:"wbskNotCheckOrigin"`    // 默认false
	GrpcHost            string            `json:"grpcHost" bson:"grpcHost" yaml:"grpcHost"`                                  // 默认本机扫描到的第一个私用IP
	GrpcPort            int               `json:"grpcPort" bson:"grpcPort" yaml:"grpcPort"`                                  // 若为空表示不启用grpc server
	GrpcKeepAlive       time.Duration     `json:"grpcKeepAlive" bson:"grpcKeepAlive" yaml:"grpcKeepAlive"`                   // 默认不启用
	GrpcCheckTimeout    string            `json:"grpcCheckTimeout" bson:"grpcCheckTimeout" yaml:"grpcCheckTimeout"`
	GrpcCheckInterval   string            `json:"grpcCheckInterval" bson:"grpcCheckInterval" yaml:"grpcCheckInterval"`
	Cache               *cache.Config     `json:"cache" bson:"cache" yaml:"cache"` // RouterConfig所用的cache
	Accesslog           *access.Config    `json:"accesslog" bson:"accesslog" yaml:"accesslog"`
	Arguments           map[string]string `json:"arguments" bson:"arguments" yaml:"arguments"`             // 默认参数
	RouterConfig        []*RouterConfig   `json:"routerConfig" bson:"routerConfig" yaml:"routerConfig"`    // 从Http的path生成相应的访问规则: proxy/plugin/cache/off
	ServerConfig        []*ServerConfig   `json:"serverConfig" bson:"serverConfig" yaml:"serverConfig"`    // 从Grpc的Service/Method生成相应http访问点的规则配置
	ServerPlugins       [][]string        `json:"serverPlugins" bson:"serverPlugins" yaml:"serverPlugins"` // 配置ServerOption
	RouterPlugins       [][]string        `json:"routerPlugins" bson:"routerPlugins" yaml:"routerPlugins"` // 配置全局的filterPlugin
}

const CKEY = "service"

func LoadConfig() (ret *Config) {
	config, ok := conf.Get(CKEY)
	if !ok {
		return
	}

	ret = new(Config)
	ret.Name, ok = conf.ElemString(config, "name")
	ret.HttpHost, ok = conf.ElemString(config, "httpHost")
	ret.HttpPort, ok = conf.ElemInt(config, "httpPort")
	ret.HttpKeepAlive, ok = conf.ElemDuration(config, "httpKeepAlive")
	ret.HttpCheckTimeout, ok = conf.ElemString(config, "httpCheckTimeout")
	ret.HttpCheckInterval, ok = conf.ElemString(config, "httpCheckInterval")
	ret.HttpCertFile, ok = conf.ElemString(config, "httpCertFile")
	ret.HttpKeyFile, ok = conf.ElemString(config, "httpKeyFile")
	ret.WbskReadBufferSize, ok = conf.ElemInt(config, "wbskReadBufferSize")
	ret.WbskWriteBufferSize, ok = conf.ElemInt(config, "wbskWriteBufferSize")
	ret.WbskNotCheckOrigin, ok = conf.ElemBool(config, "wbskNotCheckOrigin")
	ret.GrpcHost, ok = conf.ElemString(config, "grpcHost")
	ret.GrpcPort, ok = conf.ElemInt(config, "grpcPort")
	ret.GrpcKeepAlive, ok = conf.ElemDuration(config, "grpcKeepAlive")
	ret.GrpcCheckTimeout, ok = conf.ElemString(config, "grpcCheckTimeout")
	ret.GrpcCheckInterval, ok = conf.ElemString(config, "grpcCheckInterval")
	ck, ok := conf.Elem(config, "cache")
	if ok {
		ret.Cache = new(cache.Config)
		ret.Cache.Key, ok = conf.ElemString(ck, "key")
		ret.Cache.Network, ok = conf.ElemString(ck, "network")
		ret.Cache.Address, ok = conf.ElemStringSlice(ck, "address")
		ret.Cache.Keepalive, ok = conf.ElemDuration(ck, "keepalive")
		ret.Cache.ConnectTimeout, ok = conf.ElemDuration(ck, "connectTimeout")
		ret.Cache.ReadTimeout, ok = conf.ElemDuration(ck, "readTimeout")
		ret.Cache.WriteTimeout, ok = conf.ElemDuration(ck, "writeTimeout")
		ret.Cache.Password, ok = conf.ElemString(ck, "password")
		ret.Cache.InitConns, ok = conf.ElemInt(ck, "initConns")
		ret.Cache.MaxConns, ok = conf.ElemInt(ck, "maxConns")
		ret.Cache.MaxIdles, ok = conf.ElemInt(ck, "maxIdles")
		ret.Cache.TestIdleTimeout, ok = conf.ElemDuration(ck, "testIdleTimeout")
		ret.Cache.ErrExceMaxConns, ok = conf.ElemBool(ck, "errExceMaxConns")
		ret.Cache.Keyfix, ok = conf.ElemString(ck, "keyfix")
		ret.Cache.Select, ok = conf.ElemInt(ck, "select")
		ret.Cache.Cluster, ok = conf.ElemBool(ck, "cluster")
		ret.Cache.Proxyips, ok = conf.ElemStringMap(ck, "proxyips")
	}
	ac, ok := conf.Elem(config, "accesslog")
	if ok {
		ret.Accesslog = new(access.Config)
		ret.Accesslog.FlushPeriod, ok = conf.ElemDuration(ac, "flushPeriod")
		ret.Accesslog.Path, ok = conf.ElemString(ac, "path")
		ret.Accesslog.RotateBytes, ok = conf.ElemInt64(ac, "rotateBytes")
		ret.Accesslog.RotateCycle, ok = conf.ElemString(ac, "rotateCycle")
		ret.Accesslog.BufioWriterSize, ok = conf.ElemInt(ac, "bufioWriterSize")
	}
	ret.Arguments, ok = conf.ElemStringMap(config, "arguments")
	rc, ok := conf.ElemSlice(config, "routerConfig")
	if ok {
		ret.RouterConfig = make([]*RouterConfig, len(rc))
		for i, r := range rc {
			ir := new(RouterConfig)
			ir.Package, ok = conf.ElemString(r, "package")
			ir.Service, ok = conf.ElemString(r, "service")
			ir.Method, ok = conf.ElemString(r, "method")
			ir.Path, ok = conf.ElemString(r, "path")
			ir.Methods, ok = conf.ElemStringSlice(r, "methods")
			ir.ProxyPath, ok = conf.ElemString(r, "proxyPath")
			ir.ProxyService, ok = conf.ElemString(r, "proxyService")
			ir.ProxyHttps, ir.SetProxyHttps = conf.ElemBool(r, "proxyHttps")
			ps, ok := conf.ElemSlice(r, "plugins")
			if ok {
				ir.Plugins = make([][]string, len(ps))
				for i, p := range ps {
					ir.Plugins[i] = conf.ToStringSlice(p)
				}
			}
			ir.Cache, ok = conf.ElemInt64(r, "cache")
			ir.Off, ir.SetOff = conf.ElemBool(r, "off")
			ir.Access, ir.SetAccess = conf.ElemBool(r, "access")
			ret.RouterConfig[i] = ir
		}
	}
	sc, ok := conf.ElemSlice(config, "serverConfig")
	if ok {
		ret.ServerConfig = make([]*ServerConfig, len(sc))
		for i, s := range sc {
			sr := new(ServerConfig)
			sr.Package, ok = conf.ElemString(s, "package")
			sr.Service, ok = conf.ElemString(s, "service")
			sr.Method, ok = conf.ElemString(s, "method")
			sr.GrpcOff, sr.SetGrpcOff = conf.ElemBool(s, "grpcOff")
			sr.HttpOff, sr.SetHttpOff = conf.ElemBool(s, "httpOff")
			sr.HttpPath, ok = conf.ElemString(s, "httpPath")
			hps, ok := conf.ElemSlice(s, "httpPlugins")
			if ok {
				sr.HttpPlugins = make([][]string, len(hps))
				for i, p := range hps {
					sr.HttpPlugins[i] = conf.ToStringSlice(p)
				}
			}
			sr.WbskOff, sr.SetWbskOff = conf.ElemBool(s, "wbskOff")
			sr.WbskPath, ok = conf.ElemString(s, "wbskPath")
			wps, ok := conf.ElemSlice(s, "wbskPlugins")
			if ok {
				sr.WbskPlugins = make([][]string, len(wps))
				for i, p := range wps {
					sr.WbskPlugins[i] = conf.ToStringSlice(p)
				}
			}
			ret.ServerConfig[i] = sr
		}
	}
	sps, ok := conf.ElemSlice(config, "serverPlugins")
	if ok {
		ret.ServerPlugins = make([][]string, len(sps))
		for i, p := range sps {
			ret.ServerPlugins[i] = conf.ToStringSlice(p)
		}
	}
	rps, ok := conf.ElemSlice(config, "routerPlugins")
	if ok {
		ret.RouterPlugins = make([][]string, len(rps))
		for i, p := range rps {
			ret.RouterPlugins[i] = conf.ToStringSlice(p)
		}
	}
	return
}

// 合并默认值
func mergeConfig(conf *Config) *Config {

	if conf == nil {
		conf = new(Config)
	}

	// 补充默认逻辑
	if conf.HttpCheckTimeout == "" {
		conf.HttpCheckTimeout = "5s"
	}
	if conf.HttpCheckInterval == "" {
		conf.HttpCheckInterval = "6s"
	}
	if conf.GrpcCheckTimeout == "" {
		conf.GrpcCheckTimeout = "5s"
	}
	if conf.GrpcCheckInterval == "" {
		conf.GrpcCheckInterval = "6s"
	}
	return conf
}

type RouterPlugin func(args []string) gin.HandlersChain
type ServerPlugin func(args []string) grpc.ServerOption

type ServiceHandler struct {
	ServiceDesc *grpc.ServiceDesc
	ServiceImpl interface{}
	PackageName string
	ServiceName string
	Adapters    map[string]func(context.Context, []byte) (interface{}, error)
	Options     []ServiceOption
	setting     *ServiceSetting // 经过merge计算后得到的设置x
}
type RegisterServiceHandler func(service interface{}) (*grpc.ServiceDesc, string, string, map[string]func(context.Context, []byte) (interface{}, error))

type RouterSetting struct {
	PackageName  string     // package name, 支持通配符?与*
	ServiceName  string     // service name, 支持通配符?与*
	MethodName   string     // method name, 支持通配符?与*
	Path         string     // 请求路径, 支持通配符?与*
	Method       string     // 请求方法(可选,默认),多值可用逗号分隔
	ProxyPath    string     // 目标URI(必需)
	ProxyService string     // 目标服务(必需)
	ProxyHttps   bool       // 是否使用tls
	Plugins      [][]string // 基于plugin生成的过滤器
	Cache        int64      // 缓存时间(秒)
	Off          bool       // 是否关闭
	Access       bool       // 是否开启Access log, 0-关闭, 1-打印基本
}

type RouterOption func(rule *RouterSetting)

func MergeRouterConfig(configs []*RouterConfig) (ret []RouterOption) {
	for _, config := range configs {
		/* 如果是代理配置会在server.compileRounterEngine()特殊处理, 可能被替换, 也可能附加! */
		ret = append(ret, func(s *RouterSetting) {
			if (config.Package == "" || PatternMatchs(s.PackageName, config.Package)) &&
				(config.Service == "" || PatternMatchs(s.ServiceName, config.Service)) &&
				(config.Method == "" || PatternMatchs(s.MethodName, config.Method)) &&
				(config.Path == "" || PatternMatchs(s.Path, config.Path)) &&
				(len(config.Methods) == 0 || In(s.Method, config.Methods)) {

				if config.ProxyPath != "" {
					s.ProxyPath = config.ProxyPath
				}
				if config.ProxyService != "" {
					s.ProxyService = config.ProxyService
				}
				if config.ProxyHttps || config.SetProxyHttps {
					s.ProxyHttps = config.ProxyHttps
				}
				if len(config.Plugins) > 0 {
					s.Plugins = config.Plugins
				}
				if config.Cache > 0 {
					s.Cache = config.Cache
				}
				if config.Off || config.SetOff {
					s.Off = config.Off
				}
				if config.Access || config.SetAccess {
					s.Access = config.Access
				}
			}
		})
	}

	return
}

type MethodSetting struct {
	HttpOff     bool
	HttpPath    string // ServerPathDefault(packageName, serviceName, methodName) string
	HttpFilter  gin.HandlersChain
	HttpPlugins [][]string // plugins的执行次序先于filter
	WbskOff     bool
	WbskPath    string "" // ServerPathDefault(packageName, serviceName, methodName)
	WbskFilter  gin.HandlersChain
	WbskPlugins [][]string // plugins的执行次序先于filter
}

type ServiceSetting struct {
	PackageName string
	ServiceName string
	GrpcOff     bool
	Methods     map[string]*MethodSetting
}

type ServiceOption func(setting *ServiceSetting) // 会由其他方便闭包生成

func MergeServerConfig(configs []*ServerConfig) (ret []ServiceOption) {
	for _, config := range configs {
		ret = append(ret, func(s *ServiceSetting) {
			if (config.Package == "" || PatternMatchs(s.PackageName, config.Package)) &&
				(config.Service == "" || PatternMatchs(s.ServiceName, config.Service)) {
				if config.GrpcOff || config.SetGrpcOff {
					s.GrpcOff = config.GrpcOff
				}
				for m, ms := range s.Methods {
					if config.Method == "" || PatternMatchs(m, config.Method) {
						if config.HttpOff || config.SetHttpOff {
							ms.HttpOff = config.HttpOff
						}
						if config.HttpPath != "" {
							ms.HttpPath = config.HttpPath
						}
						if len(config.HttpPlugins) > 0 {
							ms.HttpPlugins = config.HttpPlugins
						}
						if config.WbskOff || config.SetWbskOff {
							ms.WbskOff = config.WbskOff
						}
						if config.WbskPath != "" {
							ms.WbskPath = config.WbskPath
						}
						if len(config.WbskPlugins) > 0 {
							ms.WbskPlugins = config.WbskPlugins
						}
					}
				}
			}
		})
	}
	return
}
