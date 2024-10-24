package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
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

		// html 渲染
		htmlTemplate *template.Template
		funcMap      template.FuncMap // 自定义渲染函数
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
	c.engine = engine
	engine.router.handle(c)

}

// Run 方法，是 ListenAndServe 的包装
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// createStaticHandler
// relativePath表示相对于当前路由组的路径，fs是一个http.FileSystem接口，用于访问文件系统中的文件。
// 首先计算出绝对路径absolutePath，它是当前路由组的前缀和传入的相对路径的组合。
// http.FileServer，并使用http.StripPrefix去除绝对路径前缀，以便正确地服务静态文件。
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 拼接URL路由路径
	absolutePath := path.Join(group.prefix, relativePath)
	// http.FileSystem 是一个接口，表示文件系统的抽象，内部实现Open方法
	// http.StripPrefix 第一个参数用于删除请求 URL 中的指定前缀，第二个参数是实际的文件路径前缀
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs)) // http.FileServer 静态文件服务器，提供目录中的文件和子目录
	return func(c *Context) {
		file := c.Param("filepath")
		log.Printf("文件路径为: %s", file)
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 文件服务器 利用createStaticHandler注册一个 GET 请求处理函数来处理静态文件请求。
func (group *RouterGroup) Static(relativePath string, root string) {
	log.Printf("当前文件路径: %s", root)
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// 注册 GET 处理程序
	group.GET(urlPattern, handler)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	// template.Must 简化模板创建的代码, 接收一个模板和一个错误
	// New 创建模板实例， Funcs 向模板添加自定义函数， ParseGlob 解析与给定模式匹配的所有模板文件，并将它们加载到模板中
	// 例如pattern = templates/* 则表示templates下所有的文件都将添加到模板实例中
	e.htmlTemplate = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}
