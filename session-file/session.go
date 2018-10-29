package session_file



import (
	. "github.com/gobigger/bigger"
	"time"
	"encoding/json"
	"github.com/tidwall/buntdb"
)



type (

	//配置文件
	fileSessionSetting struct {
		File    string
		Expiry      time.Duration
	}


	fileSessionDriver struct {}
	fileSessionConnect struct {
		config		SessionConfig
		setting		fileSessionSetting	

		db			*buntdb.DB
	}
	fileSessionValue struct {
		Value	Map
		Expiry	time.Time
	}
)









//连接
func (driver *fileSessionDriver) Connect(config SessionConfig) (SessionConnect,*Error) {

	//获取配置信息
	setting := fileSessionSetting{
		File: "store/session.db",
		Expiry: time.Hour*24*7,	//默认7天有效
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

	return &fileSessionConnect{
		config: config, setting: setting,
	},nil
}







//打开连接
func (connect *fileSessionConnect) Open() *Error {
	db,err := buntdb.Open(connect.setting.File)
	if err != nil {
		return Bigger.Erred(err)
	}
	connect.db = db
	return nil
}
func (connect *fileSessionConnect) Health() (*SessionHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &SessionHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *fileSessionConnect) Close() *Error {
	if connect.db != nil {
		if err := connect.db.Close(); err != nil {
			return Bigger.Erred(err)
		}
	}
	return nil
}







//查询会话，
func (connect *fileSessionConnect) Read(id string) (Map,*Error) {

	if connect.db == nil {
		return nil, Bigger.Erring("连接失败")
	}

	key := connect.config.Prefix + id
	val := "{}"

	err := connect.db.View(func(tx *buntdb.Tx) error {
		vvv,err := tx.Get(key)
		if err != nil {
			return err
		}
		val = vvv
		return nil
	})
	if err != nil {
		return nil, Bigger.Erred(err)
	}

	m := Map{}
	err = json.Unmarshal([]byte(val), &m)
	if err != nil {
		return nil, Bigger.Erred(err)
	} else {
		return m, nil
	}
}



//更新会话
func (connect *fileSessionConnect) Write(id string, value Map, expires ...time.Duration) *Error {
	if connect.db == nil {
		return Bigger.Erring("连接失败")
	}
	
	//JSON解析
	bytes,err := json.Marshal(value)
	if err != nil {
		return Bigger.Erred(err)
	}
	key := connect.config.Prefix + id
	val := string(bytes)

	expiry := connect.setting.Expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	err = connect.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, val, &buntdb.SetOptions{ Expires:true, TTL: expiry })
		return err
	})
	if err != nil {
		return Bigger.Erred(err)
	}

	return nil
}


//删除会话
func (connect *fileSessionConnect) Delete(id string) *Error {
	if connect.db == nil {
		return Bigger.Erring("连接失败")
	}

	//key要加上前缀
	key := connect.config.Prefix + id
	err := connect.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		return err
	})
	if err != nil {
		return Bigger.Erred(err)
	}

	return nil
}

