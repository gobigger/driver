package cache_memory


import (
	. "github.com/yatlabs/bigger"
    "sync"
	"time"
	"encoding/json"
	"github.com/tidwall/buntdb"
)






//-------------------- memoryCacheBase begin -------------------------


type (
	memoryCacheDriver struct {}
	memoryCacheConnect struct {
		mutex		sync.RWMutex
		actives		int64

		name		string
		config		CacheConfig
		setting		memoryCacheSetting

		db			*buntdb.DB
	}
	memoryCacheSetting struct {
		Expiry      time.Duration
	}

	memoryCacheBase struct {
		name		string	
		connect	*	memoryCacheConnect
		lastError	*Error
	}
	memoryCacheValue struct {
		Value	Any		`json:"value"`
	}
)











//连接
func (driver *memoryCacheDriver) Connect(name string, config CacheConfig) (CacheConnect,*Error) {
	
	//获取配置信息
	setting := memoryCacheSetting{
		Expiry: time.Hour,		//默认1小时有效
	}

	//默认超时时间
	if config.Expiry != "" {
		td,err := Bigger.Timing(config.Expiry)
		if err == nil {
			setting.Expiry = td
		}
	}
	
	return &memoryCacheConnect{
		name: name, config: config, setting: setting,
	},nil
}


//打开连接
func (connect *memoryCacheConnect) Open() *Error {
	db,err := buntdb.Open(":memory:")
	if err != nil {
		return Bigger.Erred(err)
	}
	connect.db = db
	return nil
}
func (connect *memoryCacheConnect) Health() (*CacheHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &CacheHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *memoryCacheConnect) Close() *Error {
	if connect.db != nil {
		if err := connect.db.Close(); err != nil {
			return Bigger.Erred(err)
		}
	}
	return nil
}
//获取数据库
func (connect *memoryCacheConnect) Base() (CacheBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()
	return &memoryCacheBase{connect.name, connect, nil}
}







func (base *memoryCacheBase) Close() (*Error) {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()
    return nil
}
func (base *memoryCacheBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}





func (base *memoryCacheBase) Serial(key string, nums ...int64) (int64) {
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
	} else {
		
	}

	//加数字
	value += num
	
	//写入值
	base.Write(key, value)
	if base.lastError != nil {
		return int64(0)
	}

	return value
}


//查询缓存，
func (base *memoryCacheBase) Read(key string) (Any) {
	base.lastError = nil
	
	if base.connect.db == nil {
		base.lastError = Bigger.Erring("连接失败")
		return nil
	}
	db := base.connect.db

	realKey := base.connect.config.Prefix + key
	realVal := ""

	err := db.View(func(tx *buntdb.Tx) error {
		vvv,err := tx.Get(realKey)
		if err != nil {
			return err
		}
		realVal = vvv
		return nil
	})
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return nil
	}

	mcv := memoryCacheValue{}
	err = json.Unmarshal([]byte(realVal), &mcv)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return nil
	}

	return mcv.Value
}



//更新缓存
func (base *memoryCacheBase) Write(key string, val Any, expires ...time.Duration) {
	base.lastError = nil

	if base.connect.db == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	db := base.connect.db

	//这才是值啊
	value := memoryCacheValue{val}
	
	//JSON解析
	bytes,err := json.Marshal(value)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
	
	realKey := base.connect.config.Prefix + key
	realVal := string(bytes)

	expiry := base.connect.setting.Expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	err = db.Update(func(tx *buntdb.Tx) error {
		opts := &buntdb.SetOptions{ Expires: false }
		if expiry > 0 {
			opts.Expires = true
			opts.TTL = expiry
		}
		_, _, err := tx.Set(realKey, realVal, opts)
		return err
	})
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
}

//删除缓存
func (base *memoryCacheBase) Delete(key string) {
	base.lastError = nil

	if base.connect.db == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	db := base.connect.db
	
	//key要加上前缀
	realKey := base.connect.config.Prefix + key
	err := db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(realKey)
		return err
	})
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
}

func (base *memoryCacheBase) Clear(prefixs ...string) {
	base.lastError = nil

	if base.connect.db == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	db := base.connect.db

	keys := base.Keys(prefixs...)
	if base.lastError != nil {
		return
	}
	
	err := db.Update(func(tx *buntdb.Tx) error {
		for _,key := range keys {
			_,err := tx.Delete(key)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return
	}
}
func (base *memoryCacheBase) Keys(prefixs ...string) ([]string) {
	base.lastError = nil

	keys := []string{}

	if base.connect.db == nil {
		base.lastError = Bigger.Erring("连接失败")
		return keys
	}
	db := base.connect.db
	
	err := db.View(func(tx *buntdb.Tx) error {
		if len(prefixs) >0 {
			for _,prefix := range prefixs {
				tx.AscendKeys(base.connect.config.Prefix+prefix+"*", func(k, v string) bool {
					keys = append(keys, k)
					return true
				})
			}
		} else {
			tx.AscendKeys(base.connect.config.Prefix+"*", func(k, v string) bool {
				keys = append(keys, k)
				return true
			})
		}
		
		return nil
	})
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return keys
	}

    return keys
}


//-------------------- memoryCacheBase end -------------------------