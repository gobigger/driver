package cache_memcache


import (
	. "github.com/gobigger/bigger"
    "sync"
	"time"
	"fmt"
	"github.com/pangudashu/memcache"
)






//-------------------- memcacheCacheBase begin -------------------------


type (
	memcacheCacheDriver struct {}
	memcacheCacheConnect struct {
		mutex		sync.RWMutex
		actives		int64

		name		string
		config		CacheConfig
		setting		memcacheCacheSetting

		client		*memcache.Memcache
	}
	memcacheCacheSetting struct {
		Servers     []string
		Pool		int
		Timeout		time.Duration
		Expiry      time.Duration
	}

	memcacheCacheBase struct {
		name		string	
		connect	*	memcacheCacheConnect
		lastError	*Error
	}
	memcacheCacheValue struct {
		Value	Any		`json:"value"`
	}
)











//连接
func (driver *memcacheCacheDriver) Connect(name string, config CacheConfig) (CacheConnect,*Error) {
	
	//获取配置信息
	setting := memcacheCacheSetting{
		Servers: []string{"127.0.0.1:11211"},
		Pool: 50, Timeout: time.Second*3,
		Expiry: time.Hour,		//默认1小时有效
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

	return &memcacheCacheConnect{
		name: name, config: config, setting: setting,
	},nil
}


//打开连接
func (connect *memcacheCacheConnect) Open() *Error {
	servers := []*memcache.Server{}
	for _,server := range connect.setting.Servers {
		servers = append(servers, &memcache.Server{
			Address: server, Weight: 1,
			InitConn: 15, MaxConn: connect.setting.Pool,
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
func (connect *memcacheCacheConnect) Health() (*CacheHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &CacheHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *memcacheCacheConnect) Close() *Error {
	if connect.client != nil {
		connect.client.Close()
	}
	return nil
}
//获取数据库
func (connect *memcacheCacheConnect) Base() (CacheBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()
	return &memcacheCacheBase{connect.name, connect, nil}
}




func (base *memcacheCacheBase) Close() (*Error) {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()
    return nil
}
func (base *memcacheCacheBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}





func (base *memcacheCacheBase) Serial(key string, nums ...int64) (int64) {
	base.lastError = nil

	num := int64(1)
	if len(nums) > 0 {
		num = nums[0]
	}

	value := int64(0)
	val := base.Read(key)
	if vv,ok := val.(float64); ok {
		value = int64(vv)
	} else if vv,ok := val.(int64); ok {
		value = vv
	}

	//加数字
	value += num
	
	//写入值
	base.Write(key, value, 0)
	if base.lastError != nil {
		return int64(0)
	}

	return value
}


//查询缓存，
func (base *memcacheCacheBase) Read(key string) (Any) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return nil
	}
	client := base.connect.client

	realKey := base.connect.config.Prefix + key
	realVal := memcacheCacheValue{}
	_,_,err := client.Get(realKey, &realVal)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return nil
	}

	return realVal.Value
}



//更新缓存
func (base *memcacheCacheBase) Write(key string, val Any, expires ...time.Duration) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	client := base.connect.client

	realKey := base.connect.config.Prefix + key
	realVal := memcacheCacheValue{ val }

	expiry := base.connect.setting.Expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	if expiry > 0 {
		_,err := client.Set(realKey, realVal, uint32(expiry.Seconds()))
		if err != nil {
			base.lastError = Bigger.Erred(err)
			return
		}
	} else {
		_,err := client.Set(realKey, realVal)
		if err != nil {
			base.lastError = Bigger.Erred(err)
			return
		}
	}
}

//删除缓存
func (base *memcacheCacheBase) Delete(key string) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	client := base.connect.client
	
	//key要加上前缀
	realKey := base.connect.config.Prefix + key

	_,err := client.Delete(realKey)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
}

func (base *memcacheCacheBase) Clear(prefixs ...string) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	client := base.connect.client

	keys := base.Keys(prefixs...)
	if base.lastError != nil {
		return
	}

	for _,key := range keys {
		_,err := client.Delete(key)
		if err != nil {
			base.lastError = Bigger.Erred(err)
			return
		}
	}
}
func (base *memcacheCacheBase) Keys(prefixs ...string) ([]string) {
	base.lastError = nil

	keys := []string{}

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return keys
	}
	// client := base.connect.client

	//memcache待处理返回keys

	// if err != nil {
	// 	base.lastError = Bigger.Erred(err)
	// 	return keys
	// }

    return keys
}


//-------------------- memcacheCacheBase end -------------------------