package mutex_default


import (
	. "github.com/yatlabs/bigger"
    "sync"
    "time"
)




type (
	defaultMutexDriver struct {}
	defaultMutexConnect struct {
        config      MutexConfig
        expiry      time.Duration
		mutexs    	sync.Map
	}
)


//连接
func (driver *defaultMutexDriver) Connect(config MutexConfig) (MutexConnect,*Error) {

    expiry := time.Second*12  //默认12秒有效
    if config.Expiry != "" {
        du,err := Bigger.Timing(config.Expiry)
        if err != nil {
            expiry = du
        }
    }

	return &defaultMutexConnect{
        config: config, expiry: expiry,
	}, nil
}


//打开连接
func (connect *defaultMutexConnect) Open() *Error {
	return nil
}
func (connect *defaultMutexConnect) Health() (*MutexHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &MutexHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *defaultMutexConnect) Close() *Error {
	return nil
}


func (connect *defaultMutexConnect) Lock(key string) (bool) {
	realKey := connect.config.Prefix + key
	_, exist := connect.mutexs.LoadOrStore(realKey, true)
	return exist == false
}

//删除会话
func (connect *defaultMutexConnect) Unlock(key string) (*Error) {
	realKey := connect.config.Prefix + key
	connect.mutexs.Delete(realKey)
	return nil
}

