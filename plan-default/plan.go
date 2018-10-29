package plan_default


import (
	. "github.com/gobigger/bigger"
	"github.com/gobigger/bigger/cron"
	"time"
	"sync"
	"fmt"
	"strings"
)





//------------------------- 默认计划驱动 begin --------------------------


const (
	defaultPlanSeparator = "|||"
)
type (

	defaultPlanDriver struct {}
	defaultPlanConnect struct {
		mutex		sync.RWMutex
		actives		int64

		config		PlanConfig
		handler 	PlanHandler

		cron        *cron.Cron
		
		registers	map[string]PlanRegister
		entities	map[string][]string
	}

	//响应对象
	defaultPlanResponse struct {
		connect *defaultPlanConnect
	}


	PlanData struct{
		Name    string
		Time    string
		Value   Map
	}
)







//连接
func (driver *defaultPlanDriver) Connect(config PlanConfig) (PlanConnect,*Error) {
	return &defaultPlanConnect{
		config: config,
		registers: map[string]PlanRegister{},
		entities:	map[string][]string{},
	}, nil
}


//打开连接
func (connect *defaultPlanConnect) Open() *Error {
    connect.cron = cron.New()
	return nil
}
func (connect *defaultPlanConnect) Health() (*PlanHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &PlanHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultPlanConnect) Close() *Error {
    connect.cron.Stop()
    return nil
}



//注册回调
func (connect *defaultPlanConnect) Accept(handler PlanHandler) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//保存回调
	connect.handler = handler

	return nil
}
//订阅者，注册计划
func (connect *defaultPlanConnect) Register(name string, config PlanRegister) (*Error) {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//如果已经注册，先干掉
	if exists,ok := connect.entities[name]; ok {
		for _,id := range exists {
			connect.cron.RemoveEntry(id)
		}
		delete(connect.entities, name)
	}

	ids := []string{}
	
	for i,tttt := range config.Times {
		timeName := fmt.Sprintf("%s%s%v", name, defaultPlanSeparator, i)
		id,err := connect.cron.AddFunc(tttt, func() {
			connect.serve(timeName, Map{}, config.Delay)
		}, &cron.Extra{ Name: timeName, RunForce: false, TimeOut: 5 })
	
		if err != nil {
			return Bigger.Erred(err)
		}

		ids = append(ids, id)
	}

	connect.registers[name] = config
	connect.entities[name] = ids

	return nil
}

//开始
func (connect *defaultPlanConnect) Start() *Error {
    connect.cron.Start()
	return nil
}


func (connect *defaultPlanConnect) Execute(name string, value Map) *Error {
	if _,ok := connect.registers[name]; ok {
		connect.request("", name, value, true)
	}
	return nil
}
func (connect *defaultPlanConnect) DeferredExecute(name string, delay time.Duration, value Map) *Error {
	time.AfterFunc(delay, func() {
		if _,ok := connect.registers[name]; ok {
			connect.request("", name, value, true)
		}
	})
	return nil
}



func (connect *defaultPlanConnect) serve(name string, value Map, delay bool) {
	if strings.Contains(name, defaultPlanSeparator) {
		i := strings.Index(name, defaultPlanSeparator)
		name = name[0:i]
	}

	connect.request("", name, value, delay)
}



//执行统一到这里
func (connect *defaultPlanConnect) request(id string, name string, value Map, delay bool, manuals ...bool) {
	manual := false
	if len(manuals) > 0 {
		manual = manuals[0]
	}

	req := &PlanRequest{ Id: id, Name: name, Value: value, Delay: delay, Manual: manual }
	res := &defaultPlanResponse{ connect }
	connect.handler(req, res)
}




//完成计划，从列表中清理
func (connect *defaultPlanConnect) finish(req *PlanRequest) *Error {
	//没什么要处理的
	return nil
}
//重开计划
func (connect *defaultPlanConnect) delay(req *PlanRequest, delay time.Duration) *Error {
	if req.Delay {
		time.AfterFunc(delay, func() {
			connect.request(req.Id, req.Name, req.Value, req.Delay)
		})
	}
	return nil
}





//完成计划，从列表中清理
func (res *defaultPlanResponse) Finish(req *PlanRequest) *Error {
	return res.connect.finish(req)
}
//重开计划
func (res *defaultPlanResponse) Delay(req *PlanRequest, delay time.Duration) *Error {
	return res.connect.delay(req, delay)
}




//------------------------- 默认计划驱动 end --------------------------


