package socket_default


import (
	. "github.com/yatlabs/bigger"
	"time"
	"sync"
	"net/http"
)







//------------------------- 默认事件驱动 begin --------------------------



type (

	defaultSocketDriver struct {}
	defaultSocketConnect struct {
		mutex		sync.RWMutex
		running		bool
		actives		int64

		name		string
		config		SocketConfig

		handler 	SocketHandler
		hub			*Hub
		receiver	chan *Msg
	}

	//响应对象
	defaultSocketResponse struct {
		connect *defaultSocketConnect
	}


	SocketData struct{
		Name    string
		Time    string
		Value   Map
	}
)







//连接
func (driver *defaultSocketDriver) Connect(name string, config SocketConfig) (SocketConnect,*Error) {
	return &defaultSocketConnect{
		name: name, config: config,
	}, nil
}


//打开连接
func (connect *defaultSocketConnect) Open() *Error {
	connect.hub = newHub()
	connect.receiver = make(chan *Msg)
	return nil
}
func (connect *defaultSocketConnect) Health() (*SocketHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &SocketHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultSocketConnect) Close() *Error {
	connect.hub.close()
	close(connect.receiver)
	return nil
}



//注册回调
func (connect *defaultSocketConnect) Accept(handler SocketHandler) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//保存回调
	connect.handler = handler

	return nil
}


//开始
//订阅者，发布者一起都要
func (connect *defaultSocketConnect) Start() *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	go connect.hub.run()
	go connect.receiving()
	connect.running = true

	return nil
}



func (connect *defaultSocketConnect) receiving() {
	for {
		select {
		case msg := <-connect.receiver:
			if msg != nil {
				go connect.serve(msg.id, msg.data)
			}
		}
	}
}



















func (connect *defaultSocketConnect) Upgrade(id string, req *http.Request, res http.ResponseWriter) *Error {
	if _,ok := connect.hub.clients[id]; ok {
		return Bigger.Erring("已经存在了")
	}

	ws, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		return Bigger.Erred(err)
	}

	client := &Client{
		id: id, ws: ws,
		hub: connect.hub, connect: connect,
		channels: make(map[string]bool),
		sender: make(chan []byte, 256), closer: make(chan bool),
	}
	client.hub.register <- client

	//初发器
	Bigger.Trigger(EventSocketUpgrade, Map{ "id": id })

	go client.readPump()
	client.writePump()
	client.hub.unregister <- client
	return nil
}
func (connect *defaultSocketConnect) Degrade(id string) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	if cli,ok := connect.hub.clients[id]; ok {
		cli.closer <- true
		Bigger.Trigger(EventSocketDegrade, Map{ "id": id })
		return nil
	}
	return Bigger.Erring("无效连接")
}

func (connect *defaultSocketConnect) Follow(id, channel string) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	if cli,ok := connect.hub.clients[id]; ok {
		cli.channels[channel] = true
		Bigger.Trigger(EventSocketFollow, Map{ "id": id, "channe": channel })
		return nil
	}

	return Bigger.Erring("无效连接")
}
func (connect *defaultSocketConnect) Unfollow(id, channel string) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	if cli,ok := connect.hub.clients[id]; ok {
		delete(cli.channels, channel)
		Bigger.Trigger(EventSocketUnfollow, Map{ "id": id, "channe": channel })
		return nil
	}

	return Bigger.Erring("无效连接")
}




func (connect *defaultSocketConnect) Message(id string, bytes []byte) *Error {
	if cli,ok := connect.hub.clients[id]; ok {
		cli.sender <- bytes
	}
	return nil
}
func (connect *defaultSocketConnect) DeferredMessage(id string, delay time.Duration, bytes []byte) *Error {
	time.AfterFunc(delay, func() {
		if cli,ok := connect.hub.clients[id]; ok {
			cli.sender <- bytes
		}
	})
	return nil
}
func (connect *defaultSocketConnect) Broadcast(channel string, bytes []byte) *Error {
	connect.hub.broadcast <- &Msg{ channel, bytes }
	return nil
}
func (connect *defaultSocketConnect) DeferredBroadcast(channel string, delay time.Duration, bytes []byte) *Error {
	time.AfterFunc(delay, func() {
		connect.hub.broadcast <- &Msg{ channel, bytes }
	})
	return nil
}






func (connect *defaultSocketConnect) serve(id string, data []byte) {
	connect.request(id, data)
}




//执行统一到这里
func (connect *defaultSocketConnect) request(id string, data []byte) {
	req := &SocketRequest{id, data}
	res := &defaultSocketResponse{ connect }
	connect.handler(req, res)
}


//完成事件，从列表中清理
func (connect *defaultSocketConnect) finish(req *SocketRequest) *Error {
	//没什么要处理的
	return nil
}
//重开事件
func (connect *defaultSocketConnect) delay(req *SocketRequest, delay time.Duration) *Error {
	time.AfterFunc(delay, func() {
		connect.request(req.Id, req.Data)
	})
	return nil
}





//完成事件，从列表中清理
func (res *defaultSocketResponse) Finish(req *SocketRequest) *Error {
	return res.connect.finish(req)
}
//重开事件
func (res *defaultSocketResponse) Delay(req *SocketRequest, delay time.Duration) *Error {
	return res.connect.delay(req, delay)
}




//------------------------- 默认事件驱动 end --------------------------

