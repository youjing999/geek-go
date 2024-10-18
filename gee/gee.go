package gee

import (
	"log"
	"net/http"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implements the interface of ServerHTTP
type (
	RouterGroup struct {
		prefix      string        // 路由组的前缀
		middlewares []HandlerFunc //support middleware
		parent      *RouterGroup  // 支持嵌套
		// Group对象，还需要有访问Router的能力
		// 我们可以在Group中，保存一个指针，指向Engine，整个框架的所有资源都是由Engine统一协调的
		engine *Engine
	}
	Engine struct {
		router *router
		*RouterGroup
		groups []*RouterGroup
	}
)

//New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	// 通过Engine间接地访问各种接口
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group is defined to create a new  RouterGroup
// remember all groups share the same Engine Instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add get request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Use is defined to add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

//	Engine实现的 ServeHTTP 方法的作用就是，解析请求的路径，查找路由映射表
//	如果查到，就执行注册的处理方法。如果查不到，就返回 404 NOT FOUND 。
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// middleware修改
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)

}

// Run 方法，是 ListenAndServe 的包装
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
