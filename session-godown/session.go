package session_godown



import (
	. "github.com/gobigger/bigger"
	"time"
	"encoding/json"
	"github.com/namreg/godown/client"
)



type (

	//配置文件
	godownSessionSetting struct {
		Servers     []string
		Expiry      time.Duration
	}


	godownSessionDriver struct {}
	godownSessionConnect struct {
		config		SessionConfig
		setting		godownSessionSetting	

		client		*client.Client
	}
	godownSessionValue struct {
		Value	Map
		Expiry	time.Time
	}
)









//连接
func (driver *godownSessionDriver) Connect(config SessionConfig) (SessionConnect,*Error) {

	//获取配置信息
	setting := godownSessionSetting{
		Servers: []string{"127.0.0.1:4000"}, Expiry: time.Hour*24*7,	//默认7天有效
	}

	//默认超时时间
	if config.Expiry != "" {
		td,err := Bigger.Timing(config.Expiry)
		if err == nil {
			setting.Expiry = td
		}
	}

	
	if vv,ok := config.Setting["servers"].([]string); ok {
		setting.Servers = vv
	}
	if vvs,ok := config.Setting["servers"].([]Any); ok {
		servers := []string{}
		for _,vv := range vvs {
			if vs,ok := vv.(string); ok {
				servers = append(servers, vs)
			}
		}
		setting.Servers = servers
	}
	if vv,ok := config.Setting["server"].(string); ok && vv!="" {
		setting.Servers = []string{ vv }
	}

	return &godownSessionConnect{
		config: config, setting: setting,
	},nil
}












//打开连接
func (connect *godownSessionConnect) Open() *Error {

	cli,err := client.New(connect.setting.Servers[0], connect.setting.Servers[1:]...)
	if err != nil {
		return Bigger.Erred(err)
	}

	connect.client = cli
	return nil
}
func (connect *godownSessionConnect) Health() (*SessionHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &SessionHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *godownSessionConnect) Close() *Error {
	if connect.client != nil {
		if err := connect.client.Close(); err != nil {
			return Bigger.Erred(err)
			
		}
	}
	return nil
}







//查询会话，
func (connect *godownSessionConnect) Read(id string) (Map,*Error) {

	if connect.client == nil {
		return nil, Bigger.Erring("连接失败")
	}

	key := connect.config.Prefix + id
	res := connect.client.Get(key)
	if err := res.Err(); err != nil {
		return nil, Bigger.Erred(err)
	}

	val,err := res.Val()
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
func (connect *godownSessionConnect) Write(id string, value Map, expires ...time.Duration) *Error {
	
	if connect.client == nil {
		return Bigger.Erring("连接失败")
	}
	
	//带前缀
	key := connect.config.Prefix + id

	//JSON解析
	bytes,err := json.Marshal(value)
	if err != nil {
		return Bigger.Erred(err)
	}

	expiry := connect.setting.Expiry
	if len(expires) > 0 {
		expiry = expires[0]
	}

	//写入值
	res := connect.client.Set(key, string(bytes))
	if err := res.Err(); err != nil {
		return Bigger.Erred(err)
	}

	if expiry > 0 {
		//写入过期时间
		connect.client.Expire(key, int(expiry.Seconds()))
	}
	return nil
}


//删除会话
func (connect *godownSessionConnect) Delete(id string) *Error {
	if connect.client == nil {
		return Bigger.Erring("连接失败")
	}
	
	//带前缀
	key := connect.config.Prefix + id

	//写入值
	res := connect.client.Del(key)
	if err := res.Err(); err != nil {
		return Bigger.Erred(err)
	}

	return nil
}

