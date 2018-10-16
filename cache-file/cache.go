package cache_file


import (
	. "github.com/yatlabs/bigger"
    "sync"
	"time"
	"encoding/json"
	"github.com/tidwall/buntdb"
)






//-------------------- fileCacheBase begin -------------------------


type (
	fileCacheDriver struct {}
	fileCacheConnect struct {
		mutex		sync.RWMutex
		actives		int64

		name		string
		config		CacheConfig
		setting		fileCacheSetting

		db			*buntdb.DB
	}
	fileCacheSetting struct {
		File		string
		Expiry      time.Duration
	}

	fileCacheBase struct {
		name		string	
		connect	*	fileCacheConnect
		lastError	*Error
	}
	fileCacheValue struct {
		Value	Any		`json:"value"`
	}
)











//连接
func (driver *fileCacheDriver) Connect(name string, config CacheConfig) (CacheConnect,*Error) {
	
	//获取配置信息
	setting := fileCacheSetting{
		File: "store/cache.db",
		Expiry: time.Hour,		//默认1小时有效
	}

	//默认超时时间
	if config.Expiry != "" {
		td,err := Bigger.Timing(config.Expiry)
		if err == nil {
			setting.Expiry = td
		}
	}

	if vv,ok := config.Setting["file"].(string); ok && vv!="" {
		setting.File = vv
	}
	
	return &fileCacheConnect{
		name: name, config: config, setting: setting,
	},nil
}


//打开连接
func (connect *fileCacheConnect) Open() *Error {
	db,err := buntdb.Open(connect.setting.File)
	if err != nil {
		return Bigger.Erred(err)
	}
	connect.db = db
	return nil
}
func (connect *fileCacheConnect) Health() (*CacheHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &CacheHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *fileCacheConnect) Close() *Error {
	if connect.db != nil {
		if err := connect.db.Close(); err != nil {
			return Bigger.Erred(err)
		}
	}
	return nil
}
//获取数据库
func (connect *fileCacheConnect) Base() (CacheBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()
	return &fileCacheBase{connect.name, connect, nil}
}




func (base *fileCacheBase) Close() (*Error) {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()
    return nil
}
func (base *fileCacheBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}


func (base *fileCacheBase) Serial(key string, nums ...int64) (int64) {
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
	
	//写入值，这个应该不过期
	base.Write(key, value, 0)
	if base.lastError != nil {
		return int64(0)
	}

	return value
}


//查询缓存，
func (base *fileCacheBase) Read(key string) (Any) {
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

	mcv := fileCacheValue{}
	err = json.Unmarshal([]byte(realVal), &mcv)
	if err != nil {
		base.lastError = Bigger.Erred(err)
		return nil
	}

	return mcv.Value
}



//更新缓存
func (base *fileCacheBase) Write(key string, val Any, expires ...time.Duration) {
	base.lastError = nil

	if base.connect.db == nil {
		base.lastError = Bigger.Erring("连接失败")
		return
	}
	db := base.connect.db

	value := fileCacheValue{val}
	
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
func (base *fileCacheBase) Delete(key string) {
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

func (base *fileCacheBase) Clear(prefixs ...string) {
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
func (base *fileCacheBase) Keys(prefixs ...string) ([]string) {
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


//-------------------- fileCacheBase end -------------------------