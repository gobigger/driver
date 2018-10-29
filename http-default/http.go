package http_default


import (
	. "github.com/gobigger/bigger"
	"time"
	"sync"
	"strings"
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)
















//------------------------- 默认事件驱动 begin --------------------------


const (
	defaultHttpSeparator = "|||"
)
type (
	defaultHttpDriver struct {}
	defaultHttpConnect struct {
		mutex		sync.RWMutex
		actives		int64

		config		HttpConfig

		server		*http.Server
		router		*mux.Router
		handler 	HttpHandler

		routes			map[string]*mux.Route
		registers		map[string]HttpRegister
	}

	//响应对象
	defaultHttpResponse struct {
		connect *defaultHttpConnect
	}

)










//连接
func (driver *defaultHttpDriver) Connect(config HttpConfig) (HttpConnect,*Error) {
	return &defaultHttpConnect{
		config: config,
		routes: map[string]*mux.Route{},
		registers: map[string]HttpRegister{},
	}, nil
}


//打开连接
func (connect *defaultHttpConnect) Open() *Error {
	connect.router = mux.NewRouter()
	connect.server = &http.Server{
        Addr:         	fmt.Sprintf(":%d", connect.config.Port),
        WriteTimeout:	time.Second * 15,
        ReadTimeout:	time.Second * 15,
        IdleTimeout:	time.Second * 60,
        Handler:		connect.router,
    }

	return nil
}
func (connect *defaultHttpConnect) Health() (*HttpHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &HttpHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultHttpConnect) Close() *Error {


	//安全关闭连接来

    // c := make(chan os.Signal, 1)
    // // We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
    // // SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
    // signal.Notify(c, os.Interrupt)

    // // Block until we receive our signal.
    // <-c

    // // Create a deadline to wait for.
    // ctx, cancel := context.WithTimeout(context.Background(), wait)
    // defer cancel()
    // // Doesn't block if no connections, but will otherwise wait
    // // until the timeout deadline.
    // srv.Shutdown(ctx)
    // // Optionally, you could run srv.Shutdown in a goroutine and block on
    // // <-ctx.Done() if your application should wait for other services
    // // to finalize based on context cancellation.
    // log.Println("shutting down")
    // os.Exit(0)


	return nil
}



//注册回调
func (connect *defaultHttpConnect) Accept(handler HttpHandler) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//保存回调
	connect.handler = handler

	//先注册一个接入全部请求的
	connect.router.NotFoundHandler = connect
	connect.router.MethodNotAllowedHandler = connect

	return nil
}
//订阅者，注册事件
func (connect *defaultHttpConnect) Register(name string, config HttpRegister) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//如果hosts为空，加一个空字串，用于循环，就是一定会进入循环
	if len(config.Hosts) == 0 {
		config.Hosts = append(config.Hosts, "")
	}

	//开工支持多个uri和多域名
	for i,uri := range config.Uris {
		for j,host := range config.Hosts {

			//自定义路由名称
			//Serve的时候，去掉/后面的内容，来匹配
			routeName := fmt.Sprintf("%s%s%v.%v", name, defaultHttpSeparator, i, j)
			route := connect.router.HandleFunc(uri, connect.ServeHTTP).Name(routeName)
			if len(config.Methods) > 0 {
				route.Methods(config.Methods...)
			}
			if host != "" {
				route.Host(host)
			}

			connect.routes[routeName] = route
		}
	}

	connect.registers[name] = config

	return nil
}



func (connect *defaultHttpConnect) Start() *Error {
	if connect.server == nil {
		panic("[HTTP]请先打开连接")
	}

	 go func() {
		 err := connect.server.ListenAndServe()
		 if err != nil {
			 panic(err.Error())
		 }
	 }()

	return nil
}
func (connect *defaultHttpConnect) StartTLS(certFile, keyFile string) *Error {
	if connect.server == nil {
		panic("[HTTP]请先打开连接")
	}
	 connect.server.ListenAndServeTLS(certFile, keyFile)
	return nil
}


func (connect *defaultHttpConnect) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {
	name := ""
	site := ""
	params := Map{}
	
	//有一个特别诡异的问题
	//直接从包加载的，没有问题
	//如果是从so文件加载的，这里会获取不到route
	//而先从包加载一次，再从so加载， 就可以获取到route了
	route := mux.CurrentRoute(reader)
	if route != nil {
		name = route.GetName()
		
		if strings.Contains(name, defaultHttpSeparator) {
			i := strings.Index(name, defaultHttpSeparator)
			name = name[:i]
		}

		if regis,ok := connect.registers[name]; ok {
			site = regis.Site
		}

		vars := mux.Vars(reader)
		for k,v := range vars {
			params[k] = v
		}
	}
	

	connect.request(name, site, params, writer, reader)
}

//servehttp
func (connect *defaultHttpConnect) request(name, site string, params Map, writer http.ResponseWriter, reader *http.Request) {
	if connect.handler != nil {
		req := &HttpRequest{ Name: name, Site: site, Params: params, Writer: writer, Reader: reader }
		res := &defaultHttpResponse{ connect }
		connect.handler(req, res)
	}
}






//执行统一到这里
// func (connect *defaultHttpConnect) request(id string, name string, config, value Map) {
// 	req := &HttpRequest{ Id: id, Name: name, Config: config, Value: value }
// 	res := &defaultHttpResponse{ connect }
// 	connect.handler(req, res)
// }





//完成
func (res *defaultHttpResponse) Finish(req *HttpRequest) *Error {
	return nil
}

//------------------------- 默认HTTP驱动 end --------------------------




