package queue_default


import (
	. "github.com/yatlabs/bigger"
	"time"
	"sync"
)










//------------------------- 默认队列驱动 begin --------------------------



type (

	defaultQueueMessage struct {
		mutex		sync.Mutex
		consumes	map[string]chan Map
	}
	defaultQueueFunc func(string,Map)

	defaultQueueDriver struct {}
	defaultQueueConnect struct {
		mutex		sync.RWMutex
		running		bool
		actives		int64

		name		string
		config		QueueConfig
		handler 	QueueHandler

		registers		map[string]QueueRegister
	}

	//响应对象
	defaultQueueResponse struct {
		connect *defaultQueueConnect
	}


	QueueData struct{
		Name    string
		Time    string
		Value   Map
	}
)

var (
	queueMessage *defaultQueueMessage
)
func init() {
	queueMessage = &defaultQueueMessage{ consumes: map[string]chan Map{} }
}




//订阅消息
func (msg *defaultQueueMessage) Consume(name string, regis QueueRegister, call defaultQueueFunc) (*Error) {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()

	var cc chan Map

	if vv,ok := msg.consumes[name]; ok {
		cc = vv
	} else {
		cc = make(chan Map)
		msg.consumes[name] = cc
	}

	//实际在跑的是这个
	for i:=0; i<regis.Lines; i++ {
		go msg.subscribe(cc, name, call)
	}

	return nil
}
func (msg *defaultQueueMessage) subscribe(cc chan Map, name string, call defaultQueueFunc) {
	
	//直接拉
	for {
		//这里最好使用switch，加一个关闭的通道
		//从管道获取消息
		value := <- cc
		//调用队列
		call(name, value)

	}
}





//发布消息
func (msg *defaultQueueMessage) Produce(name string, value Map) *Error {
	msg.mutex.Lock()
	defer msg.mutex.Unlock()

	//这里不能阻塞线程
	if cc,ok := msg.consumes[name]; ok {
		go func() {
			cc <- value
		}()
	}

	return nil
}









//连接
func (driver *defaultQueueDriver) Connect(name string, config QueueConfig) (QueueConnect,*Error) {
	return &defaultQueueConnect{
		name: name, config: config,
		registers: map[string]QueueRegister{},
	}, nil
}


//打开连接
func (connect *defaultQueueConnect) Open() *Error {
	return nil
}
func (connect *defaultQueueConnect) Health() (*QueueHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &QueueHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultQueueConnect) Close() *Error {
	return nil
}



//注册回调
func (connect *defaultQueueConnect) Accept(handler QueueHandler) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//保存回调
	connect.handler = handler

	return nil
}
//订阅者，注册队列
func (connect *defaultQueueConnect) Register(name string, regis QueueRegister) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	

	if connect.running == false {
		//未开始运行，保存下来
		connect.registers[name] = regis
	} else {
		//已经开始运行了，不存在才保存，而且直接订阅
		if _,ok := connect.registers[name]; ok == false {
			queueMessage.Consume(name, regis, connect.serve)
			connect.registers[name] = regis
		}
	}

	return nil
}




//开始订阅者
func (connect *defaultQueueConnect) Start() *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//订阅消息
	for name,regis := range connect.registers {
		queueMessage.Consume(name, regis, connect.serve)
	}

	connect.running = true
	return nil
}



//默认版队列这里有可能会阻塞主线程，如果lines=0就挂了
func (connect *defaultQueueConnect) Produce(name string, value Map) *Error {
	return queueMessage.Produce(name, value)
}
func (connect *defaultQueueConnect) DeferredProduce(name string, delay time.Duration, value Map) *Error {
	time.AfterFunc(delay, func() {
		queueMessage.Produce(name, value)
	})
	return nil
}



//执行统一到这里
func (connect *defaultQueueConnect) serve(name string, value Map) {
	connect.request("", name, value)
}


//执行统一到这里
func (connect *defaultQueueConnect) request(id string, name string, value Map) {
	req := &QueueRequest{ Id: id, Name: name, Value: value }
	res := &defaultQueueResponse{ connect }
	connect.handler(req, res)
}




//完成队列，从列表中清理
func (connect *defaultQueueConnect) finish(req *QueueRequest) *Error {
	//没什么要处理的
	return nil
}
//重开队列
func (connect *defaultQueueConnect) delay(req *QueueRequest, delay time.Duration) *Error {
	time.AfterFunc(delay, func() {
		connect.request(req.Id, req.Name, req.Value)
	})
	return nil
}









//完成队列，从列表中清理
func (res *defaultQueueResponse) Finish(req *QueueRequest) *Error {
	return res.connect.finish(req)
}
//重开队列
func (res *defaultQueueResponse) Delay(req *QueueRequest, delay time.Duration) *Error {
	return res.connect.delay(req, delay)
}




//------------------------- 默认队列驱动 end --------------------------
