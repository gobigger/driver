package session_default


import (
	. "github.com/gobigger/bigger"
    "sync"
    "time"
)









type (
	defaultSessionDriver struct {}
	defaultSessionConnect struct {
        config      SessionConfig
        expiry      time.Duration
		sessions    sync.Map
	}
	defaultSessionValue struct {
		Value	Map
		Expiry	time.Time
	}
)


//连接
func (driver *defaultSessionDriver) Connect(config SessionConfig) (SessionConnect,*Error) {
    expiry := time.Hour*24*7  //默认7天有效
    if config.Expiry != "" {
        du,err := Bigger.Timing(config.Expiry)
        if err != nil {
            expiry = du
        }
    }

	return &defaultSessionConnect{
        config: config, expiry: expiry,
        sessions: sync.Map{},
	},nil
}








//打开连接
func (connect *defaultSessionConnect) Open() *Error {
	return nil
}
func (connect *defaultSessionConnect) Health() (*SessionHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &SessionHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *defaultSessionConnect) Close() *Error {
	return nil
}




//查询会话，
func (connect *defaultSessionConnect) Read(id string) (Map,*Error) {
	if value,ok := connect.sessions.Load(id); ok {
		if vv,ok := value.(defaultSessionValue); ok {
			if vv.Expiry.Unix() < time.Now().Unix() {
				connect.Delete(id)
				return nil,Bigger.Erring("已过期")
			} else {
				return vv.Value,nil
			}

		} else {
			return nil,Bigger.Erring("无效会话")
		}

	} else {
		return nil,Bigger.Erring("无会话")
	}
}



//更新会话
func (connect *defaultSessionConnect) Write(id string, val Map, expires ...time.Duration) *Error {

	expiry := connect.expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	value := defaultSessionValue{
		Value: val, Expiry: time.Now().Add(expiry),
	}

	connect.sessions.Store(id, value)

	return nil
}


//删除会话
func (connect *defaultSessionConnect) Delete(id string) *Error {
	connect.sessions.Delete(id)
	return nil
}

