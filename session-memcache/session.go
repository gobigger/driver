package session_memcache



import (
	. "github.com/gobigger/bigger"
	"time"
	"fmt"
	"github.com/pangudashu/memcache"
)

type (

	//配置文件
	memcacheSessionSetting struct {
		Servers     []string
		Pool		int
		Timeout		time.Duration
		Expiry      time.Duration
	}

	memcacheSessionDriver struct {}
	memcacheSessionConnect struct {
		config		SessionConfig
		setting		memcacheSessionSetting	

		client		*memcache.Memcache
	}
	memcacheSessionValue struct {
		Value	Map
		Expiry	time.Time
	}
)


//连接
func (driver *memcacheSessionDriver) Connect(config SessionConfig) (SessionConnect,*Error) {

	//获取配置信息
	setting := memcacheSessionSetting{
		Servers: []string{"127.0.0.1:11211"},
		Pool: 100, Timeout: time.Second*3,
		Expiry: time.Hour*24*7,	//默认7天有效
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

	return &memcacheSessionConnect{
		config: config, setting: setting,
	},nil
}





//打开连接
func (connect *memcacheSessionConnect) Open() *Error {
	
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
func (connect *memcacheSessionConnect) Health() (*SessionHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &SessionHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *memcacheSessionConnect) Close() *Error {
	if connect.client != nil {
		connect.client.Close()
	}
	return nil
}



//查询会话，
func (connect *memcacheSessionConnect) Read(id string) (Map,*Error) {

	if connect.client == nil {
		return nil, Bigger.Erring("连接失败")
	}

	key := connect.config.Prefix + id
	val := Map{}
	_,_,err := connect.client.Get(key, &val)
	if err != nil {
		return nil, Bigger.Erred(err)
	}

	return val, nil

	// m := Map{}
	// err = json.Unmarshal(val.Value(), &m)
	// if err != nil {
	// 	return nil, Bigger.Erred(err)
	// } else {
	// 	return m, nil
	// }
}



//更新会话
func (connect *memcacheSessionConnect) Write(id string, value Map, expires ...time.Duration) *Error {
	
	if connect.client == nil {
		return Bigger.Erring("连接失败")
	}
	
	//带前缀
	key := connect.config.Prefix + id

	//JSON解析
	// bytes,err := json.Marshal(value)
	// if err != nil {
	// 	return Bigger.Erred(err)
	// }

	expiry := connect.setting.Expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	_,err := connect.client.Set(key, value, uint32(expiry.Seconds()))
	if err != nil {
		return Bigger.Erred(err)
	} else {
		//成功
		return nil
	}
}


//删除会话
func (connect *memcacheSessionConnect) Delete(id string) *Error {
	if connect.client == nil {
		return Bigger.Erring("连接失败")
	}

	//key要加上前缀
	key := connect.config.Prefix + id

	_,err := connect.client.Delete(key)
	if err != nil {
		return Bigger.Erred(err)
	}

	return nil
}

