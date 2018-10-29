package mutex_memcache



import (
	. "github.com/gobigger/bigger"
	"time"
	"fmt"
	"github.com/pangudashu/memcache"
)

type (

	//配置文件
	memcacheMutexSetting struct {
		Servers     []string
		Pool		int
		Timeout		time.Duration
		Expiry      time.Duration
	}

	memcacheMutexDriver struct {}
	memcacheMutexConnect struct {
		config		MutexConfig
		setting		memcacheMutexSetting	

		client		*memcache.Memcache
	}
)

//连接
func (driver *memcacheMutexDriver) Connect(config MutexConfig) (MutexConnect,*Error) {

	//获取配置信息
	setting := memcacheMutexSetting{
		Servers: []string{"127.0.0.1:11211"},
		Pool: 100, Timeout: time.Second*3,
	}

	//默认超时时间
	if config.Expiry != "" {
		td,err := Bigger.Timing(config.Expiry)
		if err == nil {
			setting.Expiry = td
		}
	}

	if vv,ok := config.Setting["servers"].([]string); ok && len(vv)>0 {
		setting.Servers = vv
	}
	if vv,ok := config.Setting["servers"].([]Any); ok && len(vv)>0 {
		servers := []string{}
		for _,vvvv := range vv {
			servers = append(servers, fmt.Sprintf("%v", vvvv))
		}
		setting.Servers = servers
	}

	if vv,ok := config.Setting["pool"].(int64); ok && vv>0 {
		setting.Pool = int(vv)
	}
	if vv,ok := config.Setting["active"].(int64); ok && vv>0 {
		setting.Pool = int(vv)
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

	return &memcacheMutexConnect{
		config: config, setting: setting,
	},nil
}





//打开连接
func (connect *memcacheMutexConnect) Open() *Error {
	
	servers := []*memcache.Server{}
	for _,server := range connect.setting.Servers {
		servers = append(servers, &memcache.Server{
			Address: server, Weight: 1,
			InitConn: 20, MaxConn: connect.setting.Pool,
		})
	}
	
    client, err := memcache.NewMemcache(servers)
	if err != nil {
		return Bigger.Erred(err)
	}

	client.SetRemoveBadServer(true)
	client.SetTimeout(connect.setting.Timeout, connect.setting.Timeout, connect.setting.Timeout)

	connect.client = client
	return nil
}
func (connect *memcacheMutexConnect) Health() (*MutexHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &MutexHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *memcacheMutexConnect) Close() *Error {
	if connect.client != nil {
		connect.client.Close()
	}
	return nil
}

//更新会话
func (connect *memcacheMutexConnect) Lock(key string) (bool) {
	
	if connect.client == nil {
		return false
	}

	//带前缀
	realKey := connect.config.Prefix + key

	expiry := connect.setting.Expiry
	// if len(expires) > 0 {
	// 	expiry = expires[0]
	// }

	_,err := connect.client.Add(realKey, true, uint32(expiry.Seconds()))
	if err != nil {
		return false
	}

	return true
}


//删除会话
func (connect *memcacheMutexConnect) Unlock(key string) *Error {
	if connect.client == nil {
		return Bigger.Erring("连接失败")
	}

	//key要加上前缀
	realKey := connect.config.Prefix + key

	_,err := connect.client.Delete(realKey)
	if err != nil {
		return Bigger.Erred(err)
	}

	return nil
}

