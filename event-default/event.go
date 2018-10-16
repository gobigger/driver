package event__default


import (
	. "github.com/yatlabs/bigger"
	"time"
	"sync"
)







//------------------------- 默认事件驱动 begin --------------------------



type (

	defaultEventMessage struct {
		mutex		sync.Mutex
		subscribes	map[string][]defaultEventFunc
	}
	defaultEventFunc func(string,Map)


	defaultEventDriver struct {}
	defaultEventConnect struct {
		mutex		sync.RWMutex
		running		bool
		actives		int64

		name		string
		config		EventConfig
		handler 	EventHandler

		registers	map[string]EventRegister
	}

	//响应对象
	defaultEventResponse struct {
		connect *defaultEventConnect
	}


	EventData struct{
		Name    string
		Time    string
		Value   Map
	}
)

var (
	eventMessage *defaultEventMessage
)
func init() {
	eventMessage = &defaultEventMessage{ subscribes: map[string][]defaultEventFunc{} }
}




//订阅消息
func (msg *defaultEventMessage) Subscribe(name string, regis EventRegister, call defaultEventFunc) *Error {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()

	if _,ok := msg.subscribes[name]; ok == false {
		msg.subscribes[name] = []defaultEventFunc{}
	}

	//加入调用列表
	msg.subscribes[name] = append(msg.subscribes[name], call)

	return nil
}



//发布消息
func (msg *defaultEventMessage) Publish(name string, value Map) *Error {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()

	if calls,ok := msg.subscribes[name]; ok {
		for _,call := range calls {
			go call(name, value)
		}
	}

	return nil
}









//连接
func (driver *defaultEventDriver) Connect(name string, config EventConfig) (EventConnect,*Error) {
	return &defaultEventConnect{
		name: name, config: config,
		registers: map[string]EventRegister{},
	}, nil
}


//打开连接
func (connect *defaultEventConnect) Open() *Error {
	return nil
}
func (connect *defaultEventConnect) Health() (*EventHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &EventHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultEventConnect) Close() *Error {
	return nil
}



//注册回调
func (connect *defaultEventConnect) Accept(handler EventHandler) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//保存回调
	connect.handler = handler

	return nil
}
//订阅者，注册事件
func (connect *defaultEventConnect) Register(name string, regis EventRegister) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	if connect.running == false {
		//未开始运行，保存下来
		connect.registers[name] = regis
	} else {
		//已经开始运行了，不存在才保存，而且直接订阅
		if _,ok := connect.registers[name]; ok == false {
			eventMessage.Subscribe(name, regis, connect.serve)
			connect.registers[name] = regis
		}
	}

	return nil
}




//开始
//订阅者，发布者一起都要
func (connect *defaultEventConnect) Start() *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//订阅消息
	for name,regis := range connect.registers {
		eventMessage.Subscribe(name, regis, connect.serve)
	}

	connect.running = true

	return nil
}






func (connect *defaultEventConnect) Trigger(name string, value Map) *Error {
	go connect.serve(name, value)
	return nil
}
func (connect *defaultEventConnect) SyncTrigger(name string, value Map) *Error {
	connect.serve(name, value)
	return nil
}
func (connect *defaultEventConnect) Publish(name string, value Map) *Error {
	return eventMessage.Publish(name, value)
}
func (connect *defaultEventConnect) DeferredPublish(name string, delay time.Duration, value Map) *Error {
	time.AfterFunc(delay, func() {
		eventMessage.Publish(name, value)
	})
	return nil
}






func (connect *defaultEventConnect) serve(name string, value Map) {
	connect.request("", name, value)
}


//执行统一到这里
func (connect *defaultEventConnect) request(id string, name string, value Map) {
	req := &EventRequest{ Id: id, Name: name, Value: value }
	res := &defaultEventResponse{ connect }
	connect.handler(req, res)
}


//完成事件，从列表中清理
func (connect *defaultEventConnect) finish(req *EventRequest) *Error {
	//没什么要处理的
	return nil
}
//重开事件
func (connect *defaultEventConnect) delay(req *EventRequest, delay time.Duration) *Error {
	time.AfterFunc(delay, func() {
		connect.request(req.Id, req.Name, req.Value)
	})
	return nil
}





//完成事件，从列表中清理
func (res *defaultEventResponse) Finish(req *EventRequest) *Error {
	return res.connect.finish(req)
}
//重开事件
func (res *defaultEventResponse) Delay(req *EventRequest, delay time.Duration) *Error {
	return res.connect.delay(req, delay)
}




//------------------------- 默认事件驱动 end --------------------------

