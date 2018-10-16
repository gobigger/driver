package cache_default


import (
	. "github.com/yatlabs/bigger"
    "sync"
	"time"
	"fmt"
	"strings"
)






//-------------------- defaultCacheBase begin -------------------------


type (
	defaultCacheDriver struct {}
	defaultCacheConnect struct {
		mutex		sync.RWMutex
		actives		int64
		name		string
		config		CacheConfig
		caches		sync.Map
	}
	defaultCacheBase struct {
		name		string	
		connect	*	defaultCacheConnect
		lastError	*Error
	}
	defaultCacheValue struct {
		Value	Any
		Expiry	*time.Time
	}
)











//连接
func (driver *defaultCacheDriver) Connect(name string, config CacheConfig) (CacheConnect,*Error) {
	return &defaultCacheConnect{
		name: name, config: config,
		caches: sync.Map{}, actives: int64(0),
	}, nil
}


//打开连接
func (connect *defaultCacheConnect) Open() *Error {
	return nil
}
func (connect *defaultCacheConnect) Health() (*CacheHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &CacheHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultCacheConnect) Close() *Error {
	return nil
}
//获取数据库
func (connect *defaultCacheConnect) Base() (CacheBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()
	return &defaultCacheBase{connect.name, connect, nil}
}







func (base *defaultCacheBase) Close() (*Error) {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()
    return nil
}
func (base *defaultCacheBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}



func (base *defaultCacheBase) Serial(key string, nums ...int64) (int64) {
	base.lastError = nil

	num := int64(1)
	if len(nums) > 0 {
		num = nums[0]
	}

	value := int64(0)
	if vv,ok := base.Read(key).(int64); ok {
		value = vv
	}

	value += num
	
	//写入值
	base.Write(key, value)
	if err := base.Erred(); err != nil {
		base.lastError = err
		value = int64(0)
	}
	
	return value
}


//查询缓存，
func (base *defaultCacheBase) Read(id string) (Any) {
	base.lastError = nil
	var read Any

	if value,ok := base.connect.caches.Load(id); ok {
		if vv,ok := value.(defaultCacheValue); ok {
			if vv.Expiry != nil && vv.Expiry.Unix() < time.Now().Unix() {
				base.Delete(id)
				base.lastError = Bigger.Erring("已过期")
			} else {
				read = vv.Value
			}

		} else {
			base.lastError = Bigger.Erring("无效缓存")
		}

	} else {
		base.lastError = Bigger.Erring("无缓存")
	}

	return read
}



//更新缓存
func (base *defaultCacheBase) Write(key string, val Any, expires ...time.Duration) {
	base.lastError = nil

	value := defaultCacheValue{
		Value: val,
    }
    if len(expires) > 0 {
        tm := time.Now().Add(expires[0])
        value.Expiry = &tm
    }

	base.connect.caches.Store(key, value)
}


//删除缓存
func (base *defaultCacheBase) Delete(key string) {
	base.lastError = nil
	base.connect.caches.Delete(key)
}

func (base *defaultCacheBase) Clear(prefixs ...string) {
	base.lastError = nil
	keys := base.Keys(prefixs...)
	for _,key := range keys {
		base.connect.caches.Delete(key)
	}
}
func (base *defaultCacheBase) Keys(prefixs ...string) ([]string) {
	base.lastError = nil

	keys := []string{}
	base.connect.caches.Range(func(k, v Any) bool {
		key := fmt.Sprintf("%v", k)

		if len(prefixs) == 0 {
			keys = append(keys, key)
		} else {
			for _,pre := range prefixs {
				if strings.HasPrefix(key, pre) {
					keys = append(keys, key)
					break
				}
			}
		}
		return true
	})
    return keys
}


//-------------------- defaultCacheBase end -------------------------