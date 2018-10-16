package cache_redis


import (
	. "github.com/yatlabs/bigger"
    "sync"
	"time"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
)






//-------------------- redisCacheBase begin -------------------------


type (
	redisCacheDriver struct {}
	redisCacheConnect struct {
		mutex		sync.RWMutex
		actives		int64

		name		string
		config		CacheConfig
		setting		redisCacheSetting

		client		*redis.Pool
	}
	redisCacheSetting struct {
		Server      string      //服务器地址，ip:端口
		Password    string      //服务器auth密码
		Database    string      //数据库
		Expiry      time.Duration

		Idle        int         	//最大空闲连接
		Active      int         	//最大激活连接，同时最大并发
		Timeout     time.Duration
	}

	redisCacheBase struct {
		name		string	
		connect	*	redisCacheConnect
		lastError	*Error
	}
	redisCacheValue struct {
		Value	Any		`json:"value"`
	}
)











//连接
func (driver *redisCacheDriver) Connect(name string, config CacheConfig) (CacheConnect,*Error) {
	
	//获取配置信息
	setting := redisCacheSetting{
		Server: "127.0.0.1:6379", Password: "", Database: "",
		Idle: 30, Active: 100, Timeout: 240,
		Expiry: time.Hour,		//默认1小时有效
	}

	//默认超时时间
	if config.Expiry != "" {
		td,err := Bigger.Timing(config.Expiry)
		if err == nil {
			setting.Expiry = td
		}
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

	return &redisCacheConnect{
		name: name, config: config, setting: setting,
	},nil
}


//打开连接
func (connect *redisCacheConnect) Open() *Error {
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
func (connect *redisCacheConnect) Health() (*CacheHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &CacheHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *redisCacheConnect) Close() *Error {
	if connect.client != nil {
		if err := connect.client.Close(); err != nil {
			return Bigger.Erred(err)
			
		}
	}
	return nil
}
//获取数据库
func (connect *redisCacheConnect) Base() (CacheBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()
	return &redisCacheBase{connect.name, connect, nil}
}

















func (base *redisCacheBase) Close() (*Error) {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()
    return nil
}
func (base *redisCacheBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}





func (base *redisCacheBase) Serial(key string, nums ...int64) (int64) {
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
func (base *redisCacheBase) Read(key string) (Any) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return nil
	}
	conn := base.connect.client.Get()
	defer conn.Close()

	realKey := base.connect.config.Prefix + key
	

	strVal,err := redis.String(conn.Do("GET", realKey))
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return nil
	}

	realVal := redisCacheValue{}
	err = json.Unmarshal([]byte(strVal), &realVal)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return nil
	}

	return realVal.Value
}



//更新缓存
func (base *redisCacheBase) Write(key string, val Any, expires ...time.Duration) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	conn := base.connect.client.Get()
	defer conn.Close()

	realKey := base.connect.config.Prefix + key
	realVal := redisCacheValue{ val }

	expiry := base.connect.setting.Expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	bytes,err := json.Marshal(realVal)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}

	args := []Any{
		realKey, string(bytes),
	}
	if expiry > 0 {
		args = append(args, "EX", expiry.Seconds())
	}

	_,err = conn.Do("SET", args...)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
}

//删除缓存
func (base *redisCacheBase) Delete(key string) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	conn := base.connect.client.Get()
	defer conn.Close()

	realKey := base.connect.config.Prefix + key

	_,err := conn.Do("DEL", realKey)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
}

func (base *redisCacheBase) Clear(prefixs ...string) {
	base.lastError = nil

	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	conn := base.connect.client.Get()
	defer conn.Close()

	keys := base.Keys(prefixs...)
	if base.lastError != nil {
		return
	}

	for _,key := range keys {
		_,err := conn.Do("DEL", key)
		if err != nil {
			base.lastError = Bigger.Erred(err)
			return
		}
	}
}
func (base *redisCacheBase) Keys(prefixs ...string) ([]string) {
	base.lastError = nil

	keys := []string{}

	
	if base.connect.client == nil {
		base.lastError = Bigger.Erring("连接失败")
		return keys
	}
	conn := base.connect.client.Get()
	defer conn.Close()

	if len(prefixs) >0 {
		for _,prefix := range prefixs {
			alls,_ := redis.Strings(conn.Do("KEYS", base.connect.config.Prefix+prefix+"*"))
			for _,key := range alls {
				keys = append(keys, key)
			}
		}
	} else {
		alls,_ := redis.Strings(conn.Do("KEYS", base.connect.config.Prefix+"*"))
		for _,key := range alls {
			keys = append(keys, key)
		}
	}

    return keys
}


//-------------------- redisCacheBase end -------------------------