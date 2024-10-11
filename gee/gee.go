package gee

import (
	"net/http"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implements the interface of ServerHTTP
type Engine struct {
	router *router
}

//New is the constructor of gee.Engine
func New() *Engine {
	return &Engine{
		router: newRouter(),
	}
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add get request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

//	Engine实现的 ServeHTTP 方法的作用就是，解析请求的路径，查找路由映射表
//	如果查到，就执行注册的处理方法。如果查不到，就返回 404 NOT FOUND 。
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)

}

// Run 方法，是 ListenAndServe 的包装
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}
