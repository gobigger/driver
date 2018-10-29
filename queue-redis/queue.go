package queue_redis


import (
	. "github.com/gobigger/bigger"
	"time"
	"sync"
	"encoding/json"
	"strings"
	"github.com/gomodule/redigo/redis"
)







//------------------------- 默认队列驱动 begin --------------------------



type (

	redisQueueDriver struct {}
	redisQueueConnect struct {
		mutex		sync.RWMutex
		running		bool
		actives		int64

		name		string
		config		QueueConfig
		setting		redisQueueSetting

		handler 	QueueHandler
		registers	map[string]QueueRegister
	
		client		*redis.Pool
		reload		string
	}
	redisQueueSetting struct {
		Server      string      //服务器地址，ip:端口
		Password    string      //服务器auth密码
		Database    string      //数据库

		Idle        int         	//最大空闲连接
		Active      int         	//最大激活连接，同时最大并发
		Timeout     time.Duration
	}

	//响应对象
	redisQueueResponse struct {
		connect *redisQueueConnect
	}

	QueueData struct{
		Name    string
		Time    string
		Value   Map
	}
)



//连接
func (driver *redisQueueDriver) Connect(name string, config QueueConfig) (QueueConnect,*Error) {

	//获取配置信息
	setting := redisQueueSetting{
		Server: "127.0.0.1:6379", Password: "", Database: "",
		Idle: 30, Active: 100, Timeout: 240,
	}

	
	if vv,ok := config.Setting["server"].(string); ok && vv!="" {
		setting.Server = vv
	}
	if vv,ok := config.Setting["password"].(string); ok && vv!="" {
		setting.Password = vv
	}

	//数据库，redis的0-16号
	if v,ok := config.Setting["database"].(string); ok {
		setting.Database = v
	}
	
	if vv,ok := config.Setting["idle"].(int64); ok && vv>0 {
		setting.Idle = int(vv)
	}
	if vv,ok := config.Setting["active"].(int64); ok && vv>0 {
		setting.Active = int(vv)
	}
	if vv,ok := config.Setting["timeout"].(int64); ok && vv>0 {
		setting.Timeout = time.Second*time.Duration(vv)
	}
	if vv,ok := config.Setting["timeout"].(string); ok && vv!="" {
		td,err := Bigger.Timing(vv)
		if err == nil {
			setting.Timeout = td
		}
	}
	
	reload := Bigger.Unique()
	return &redisQueueConnect{
		name: name, config: config, setting: setting,
		reload: reload, registers: map[string]QueueRegister{},
	}, nil
}


//打开连接
func (connect *redisQueueConnect) Open() *Error {
	connect.client = &redis.Pool{
		MaxIdle: connect.setting.Idle, MaxActive: connect.setting.Active, IdleTimeout: connect.setting.Timeout,
		Dial: func () (redis.Conn, error) {
			c, err := redis.Dial("tcp", connect.setting.Server)
			if err != nil {
				Bigger.Warning("session.redis.dial", err)
				return nil, err
			}

			//如果有验证
			if connect.setting.Password != "" {
				if _, err := c.Do("AUTH", connect.setting.Password); err != nil {
					c.Close()
					Bigger.Warning("session.redis.auth", err)
					return nil, err
				}
			}
			//如果指定库
			if connect.setting.Database != "" {
				if _, err := c.Do("SELECT", connect.setting.Database); err != nil {
					c.Close()
					Bigger.Warning("session.redis.select", err)
					return nil, err
				}
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	//打开一个试一下
	conn := connect.client.Get(); defer conn.Close()
	if err := conn.Err(); err != nil {
		return Bigger.Erred(err)
	}
	return nil
}
func (connect *redisQueueConnect) Health() (*QueueHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &QueueHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *redisQueueConnect) Close() *Error {
	if connect.client != nil {
		if err := connect.client.Close(); err != nil {
			return Bigger.Erred(err)
			
		}
	}
	return nil
}




//注册回调
func (connect *redisQueueConnect) Accept(handler QueueHandler) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//保存回调
	connect.handler = handler

	return nil
}
//订阅者，注册队列
func (connect *redisQueueConnect) Register(name string, regis QueueRegister) *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	if connect.running == false {
		//未开始运行，保存下来
		connect.registers[name] = regis
	} else {

		//已经开始运行了，不存在才保存，而且直接订阅
		if _,ok := connect.registers[name]; ok == false {
			connect.registers[name] = regis
			connect.Produce(connect.reload, Map{})
		}
	}

	return nil
}




//开始
//订阅者，发布者一起都要
func (connect *redisQueueConnect) Start() *Error {
	connect.mutex.Lock()
	defer connect.mutex.Unlock()

	//订阅消息
	go connect.consuming()

	connect.running = true

	return nil
}



//这里是调用的，要一直循环啊，连接一直不关的样子
// func (connect *redisQueueConnect) consume()  {
// 	for name,regis := range connect.registers {
// 		for i:=0; i<regis.Lines; i ++ {
// 			go connect.consuming(name)
// 		}
// 	}
// }


//这里是调用的，要一直循环啊，连接一直不关的样子
func (connect *redisQueueConnect) consuming()  {
	names := []Any{
		connect.reload,
	}
	for name,regis := range connect.registers {
		for i:=0; i<regis.Lines; i++ {
			names = append(names, connect.config.Prefix+name)
		}
	}

	//20秒超时
	names = append(names, 20)

	conn := connect.client.Get()
	defer conn.Close()

	for {
		strs,_ := redis.Strings(conn.Do("BRPOP", names...))
		if strs != nil && len(strs) >= 2 {
			if strs[0] == connect.reload {
				break
			} else {
				go connect.serve(strs[0], strs[1])
			}
		}
	}

	//递归
	connect.consuming()
}

// func (connect *redisQueueConnect) consuming(name string) {
// 	conn := connect.client.Get()
// 	defer conn.Close()

// 	realName := connect.config.Prefix + name
// 	for {
// 		bytes,err := redis.Bytes(conn.Do("BRPOP", realName, 20))
// 		Bigger.Debug("pppoppop")
// 		if err == nil {
// 			connect.serve(name, bytes)
// 		}
// 	}
// }

func (connect *redisQueueConnect) serve(name string, val string) {
	name = strings.Replace(name, connect.config.Prefix, "", 1)

	value := Map{}
	json.Unmarshal([]byte(val), &value)

	connect.request("", name, value)
}






func (connect *redisQueueConnect) Produce(name string, value Map) *Error {
	//再转成json
	bytes, err := json.Marshal(value)
	if err != nil {
		return Bigger.Erred(err)
	}

	if connect.client == nil {
		return Bigger.Erring("无效失败")
	}
	conn := connect.client.Get()
	defer conn.Close()

	//写入
	realName := connect.config.Prefix + name
	_,err = conn.Do("LPUSH", realName, string(bytes))
	if err != nil {
		return Bigger.Erred(err)
	}
	return nil
}
func (connect *redisQueueConnect) DeferredProduce(name string, delay time.Duration, value Map) *Error {
	time.AfterFunc(delay, func() {
		connect.Produce(name, value)
	})
	return nil
}



//执行统一到这里
func (connect *redisQueueConnect) request(id string, name string, value Map) {
	req := &QueueRequest{ Id: id, Name: name, Value: value }
	res := &redisQueueResponse{ connect }
	connect.handler(req, res)
}


//完成队列，从列表中清理
func (connect *redisQueueConnect) finish(req *QueueRequest) *Error {
	//没什么要处理的
	return nil
}
//重开队列
func (connect *redisQueueConnect) delay(req *QueueRequest, delay time.Duration) *Error {
	time.AfterFunc(delay, func() {
		connect.request(req.Id, req.Name, req.Value)
	})
	return nil
}



//完成队列，从列表中清理
func (res *redisQueueResponse) Finish(req *QueueRequest) *Error {
	return res.connect.finish(req)
}
//重开队列
func (res *redisQueueResponse) Delay(req *QueueRequest, delay time.Duration) *Error {
	return res.connect.delay(req, delay)
}



//------------------------- 默认队列驱动 end --------------------------

