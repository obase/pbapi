package pbapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	MethodGet        = http.MethodGet
	MethodHead       = http.MethodHead
	MethodPost       = http.MethodPost
	MethodPut        = http.MethodPut
	MethodPatch      = http.MethodPatch
	MethodDelete     = http.MethodDelete
	MethodConnect    = http.MethodConnect
	MethodOptions    = http.MethodOptions
	MethodTrace      = http.MethodTrace
	MethodStatic     = "Static"
	MethodStaticFile = "StaticFile"
	MethodStaticFS   = "StaticFS"
)

type Node struct {
	PackageName string
	ServiceName string
	MethodName  string
	Filter      gin.HandlersChain // 必须采取复制形式
	Handler     gin.HandlerFunc   // 处理器, 一般只会缓存处理器的内容
	File        string            // 文件或目录
	FileSystem  http.FileSystem   // 文件系统
}

type FlatNode struct {
	// 克隆Node的属性信息
	PackageName string
	ServiceName string
	MethodName  string
	Filter      gin.HandlersChain // 必须采取复制形式
	Handler     gin.HandlerFunc   // 处理器, 一般只会缓存处理器的内容
	File        string            // 文件或目录
	FileSystem  http.FileSystem   // 文件系统
	// 补全Node的索引信息
	Path   string
	Method string

	// 补全RouterSetting信息,最后才一并处理
	Access  bool
	Off     bool
	Cache   int64
	Plugins [][]string
}

/*
扩展gin.Engine:
1. 缓存所有gin.HandlerFunc
2. 整合httpCache, httpPlugin等扩展功能
3. 对等转换为gin.Engine
*/
type Router struct {
	path    string
	filters gin.HandlersChain
	handler map[string]map[string]*Node
	child   []*Router
}

func newRouter(path string, use gin.HandlersChain) *Router {
	return &Router{
		path:    path,
		filters: use,
		handler: make(map[string]map[string]*Node),
		child:   nil,
	}
}

func (r *Router) Group(path string, use ...gin.HandlerFunc) *Router {
	sr := newRouter(path, use)
	r.child = append(r.child, sr)
	return sr
}

func (r *Router) Use(fs ...gin.HandlerFunc) *Router {
	r.filters = append(r.filters, fs...)
	return r
}

func (r *Router) handle(method string, path string, n *Node) *Router {
	mnodes, ok := r.handler[method]
	if !ok {
		mnodes = make(map[string]*Node)
		r.handler[method] = mnodes
	} else if v, ok := mnodes[path]; ok {
		panic(fmt.Sprintf("handle conflict path: %v, new: %v, old: %v", path, n, v))
	}
	mnodes[path] = n
	return r
}

func (r *Router) Handle(method string, path string, fs ...gin.HandlerFunc) *Router {

	// 没有检测,要求fs不能为空
	last := len(fs) - 1
	n := new(Node)
	n.Handler = fs[last]
	if last > 0 {
		n.Filter = fs[:last]
	}
	r.handle(method, path, n)

	return r
}
func (r *Router) Any(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodGet, path, f...)
	r.Handle(http.MethodPost, path, f...)
	r.Handle(http.MethodDelete, path, f...)
	r.Handle(http.MethodPatch, path, f...)
	r.Handle(http.MethodPut, path, f...)
	r.Handle(http.MethodOptions, path, f...)
	r.Handle(http.MethodHead, path, f...)
	return r
}
func (r *Router) GET(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodGet, path, f...)
	return r
}
func (r *Router) POST(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodPost, path, f...)
	return r
}
func (r *Router) DELETE(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodDelete, path, f...)
	return r
}
func (r *Router) PATCH(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodPatch, path, f...)
	return r
}
func (r *Router) PUT(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodPut, path, f...)
	return r
}
func (r *Router) OPTIONS(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodOptions, path, f...)
	return r
}
func (r *Router) HEAD(path string, f ...gin.HandlerFunc) *Router {
	r.Handle(http.MethodHead, path, f...)
	return r
}

func (r *Router) StaticFile(path string, file string) *Router {
	return r.handle(MethodStaticFile, path, &Node{
		File: file,
	})
}
func (r *Router) Static(path string, file string) *Router {
	return r.handle(MethodStatic, path, &Node{
		File: file,
	})
}
func (r *Router) StaticFS(path string, fs http.FileSystem) *Router {
	return r.handle(MethodStaticFS, path, &Node{
		FileSystem: fs,
	})
}

func (r *Router) Flattern() (ret []*FlatNode) {
	_flatten(&ret, r, "", nil)
	return
}

func _flatten(ret *[]*FlatNode, parent *Router, prefix string, chain gin.HandlersChain) {
	if len(parent.path) > 0 {
		prefix = prefix + parent.path
	}
	if len(parent.filters) > 0 {
		chain = append(chain, parent.filters...)
	}
	for method, handleMap := range parent.handler {
		for path, node := range handleMap {
			fnode := &FlatNode{
				PackageName: node.PackageName,
				ServiceName: node.ServiceName,
				MethodName:  node.MethodName,
				Filter:      node.Filter,
				Handler:     node.Handler,
				File:        node.File,
				FileSystem:  node.FileSystem,
				Path:        prefix + path,
				Method:      method,
			}
			// 组装前置的filter
			if len(chain) > 0 {
				fnode.Filter = joinHandlersChain(chain, fnode.Filter)
			}
			// 组装前置的path
			*ret = append(*ret, fnode)
		}
	}

	// 递归遍历子结点
	for _, child := range parent.child {
		_flatten(ret, child, prefix, chain)
	}
}

func joinHandlersChain(c1 gin.HandlersChain, c2 gin.HandlersChain) gin.HandlersChain {
	ln1, ln2 := len(c1), len(c2)
	if ln2 == 0 {
		return c1
	}
	if ln1 == 0 {
		return c2
	}
	ret := make(gin.HandlersChain, ln1+ln2)
	copy(ret, c1)
	copy(ret[ln1:], c2)
	return ret
}
