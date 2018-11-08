package file_default


import (
	. "github.com/gobigger/bigger"
	"strings"
    "sync"
	"fmt"
	"os"
	"io"
	"path"
	"time"
	"github.com/gobigger/bigger/hashring"
	"github.com/disintegration/imaging"
)



//-------------------- defaultFileBase begin -------------------------

type (
	defaultFileDriver struct {}
	defaultFileConnect struct {
		mutex		sync.RWMutex
		actives		int64

		name		string
		config		FileConfig
		setting		defaultFileSetting

		hashring	*hashring.HashRing
	}
	defaultFileSetting struct {
		Sharding	int
		Expiry		time.Duration
		Storage		string
		Thumbnail	string
	}

	defaultFileBase struct {
		name		string	
		connect	*	defaultFileConnect
		lastError	*Error
	}
)

//连接
func (driver *defaultFileDriver) Connect(name string, config FileConfig) (FileConnect,*Error) {

	setting := defaultFileSetting{
		Sharding: 1200, Expiry: time.Hour*24,
		Storage: "store/storage", Thumbnail: "store/thumbnail",
	}

	//默认超时时间
	if config.Expiry != "" {
		td,err := Bigger.Timing(config.Expiry)
		if err == nil {
			setting.Expiry = td
		}
	}

	if vv,ok := config.Setting["sharding"].(int64); ok && vv > 0 {
		setting.Sharding = int(vv)
	}
	if vv,ok := config.Setting["storage"].(string); ok && vv != "" {
		setting.Storage = vv
	}
	if vv,ok := config.Setting["thumbnail"].(string); ok && vv != "" {
		setting.Thumbnail = vv
	}




	//分散目录数应该可以放到setting里自定义
	//一旦确认就不能修改，因为修改后文件被重新分散，老文件就会无法访问
	rings := map[string]int{}
	for i:=1;i<=setting.Sharding;i++ {
		rings[fmt.Sprintf("%v", i)] = 1
	}

	return &defaultFileConnect{
		actives: int64(0),
		name: name, config: config, setting: setting,
		hashring: hashring.New(rings),
	}, nil

}


//打开连接
func (connect *defaultFileConnect) Open() *Error {
	return nil
}
func (connect *defaultFileConnect) Health() (*FileHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &FileHealth{ Workload: connect.actives },nil
}
//关闭连接
func (connect *defaultFileConnect) Close() *Error {
	return nil
}
//获取数据库
func (connect *defaultFileConnect) Base() (FileBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()
	return &defaultFileBase{connect.name, connect, nil}
}







func (base *defaultFileBase) Close() (*Error) {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()
    return nil
}
func (base *defaultFileBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}

//默认存储，id直接记录文件路径，这样靠谱
//但是这样的话，缩略图不好生成，因为拿不到ID什么的当目录
//还是需要多个参数， 组合一个字串， 然后base64来生成id
//数据格式，使用之后就不可以改了，或者把格式也编入字串
//另外有可能int64,JSON解析会掉精度，数字必须保存成字符串
//base||hashring||id||extension
//可以考虑用hash分布
func (base *defaultFileBase) Assign(name string, metadata Map) (string) {
	base.lastError = nil

	if name=="" {
		base.lastError = Bigger.Erring("无效数据")
		return ""
	}
	//要不要判断一下文件是否已经存在？

	ffff := ""

	return Bigger.Encode(base.name, name, ffff)
}
func (base *defaultFileBase) Storage(code string, reader io.Reader) int64 {
	base.lastError = nil

	//解析ID串，拿到数据
	data := Bigger.Decode(code)
	if data == nil {
		base.lastError = Bigger.Erring("解析失败")
		return 0
	}

	_, _, sFile, err := base.storaging(data)
	if err != nil {
		base.lastError = err
		return 0
	}

	//创建文件
	save, erro := os.OpenFile(sFile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		base.lastError = Bigger.Erred(erro)
		return 0
	}
	defer save.Close()

	//复制文件
	n,erro := io.Copy(save, reader)
	if err != nil {
		base.lastError = Bigger.Erred(erro)
		return 0
	}

	return n
}
func (base *defaultFileBase) Download(key string) (io.ReadCloser, *FileCoding) {
	
	//解析ID串，拿到数据
	data := Bigger.Decode(key)

	_, _, sFile, err := base.storaging(data)
	if err != nil {
		base.lastError = err
		return nil, nil
	}

	//打开文件
	reader,erro := os.Open(sFile)
	if erro != nil {
		base.lastError = Bigger.Erred(erro)
		return nil, nil
	}

	return reader, data
}
func (base *defaultFileBase) Thumbnail(key string, width, height, num int64) (io.ReadCloser, *FileCoding) {
	//解析ID串，拿到数据
	data := Bigger.Decode(key)
	if data == nil {
		base.lastError = Bigger.Erring("解析失败")
		return nil, nil
	}

	//先获取缩略图的文件
	_,_,tfile,err := base.thumbnailing(data, width, height, num)
	if err != nil {
		base.lastError = err
		return nil, nil
	}

	//如果缩图已经存在了，就直接返回
	exist,erro := os.Open(tfile)
	if erro == nil {
		return exist, data
	}
	exist.Close()

	//获取存储的文件
	_,_,sfile,err := base.storaging(data)
	if err != nil {
		base.lastError = err
		return nil, nil
	}

	sf,erro := os.Open(sfile)
	if erro != nil {
		base.lastError = Bigger.Erred(erro)
		return nil, nil
	}
	defer sf.Close()
	
	img,erro := imaging.Decode(sf)
	if erro != nil {
		base.lastError = Bigger.Erred(erro)
		return nil, nil
	}
	
	thumb := imaging.Thumbnail(img, int(width), int(height), imaging.NearestNeighbor)
	erro = imaging.Save(thumb, tfile)
	if erro != nil {
		base.lastError = Bigger.Erred(erro)
		return nil, nil
	}

	
	//重新打开
	tf,erro := os.Open(tfile)
	if erro != nil {
		base.lastError = Bigger.Erred(erro)
		return nil, nil
	}

	return tf, data

}


func (base *defaultFileBase) Delete(key string) (*FileCoding) {
	//解析ID串，拿到数据
	data := Bigger.Decode(key)

	_, _, sFile, err := base.storaging(data)
	if err != nil {
		base.lastError = err
		return nil
	}

	erro := os.Remove(sFile)
	if err != nil {
		base.lastError = Bigger.Erred(erro)
		return nil
	}

	return data
}

func (base *defaultFileBase) Browse(code string, name string, args Map, expires ...time.Duration) (string) {
	expiry := time.Hour*24

	if base.connect.config.Expiry != "" {
		if vv,err := Bigger.Timing(base.connect.config.Expiry); err == nil {
			expiry = vv
		}
	}
	if len(expires) > 0 {
		expiry = expires[0]
	}

	//ip, expiry, name
	deadline := time.Now().Add(expiry)
	tokens := []string{
		fmt.Sprintf("%v", args["ip"]),
		fmt.Sprintf("%d", deadline.UnixNano()),
	}
	data := Bigger.Decode(code)
	if data != nil {
		tokens = append(tokens, data.Name)
	}

	token := Bigger.Encrypts(tokens)

	browse := browseName
	if base.connect.config.Browse != "" {
		browse = base.connect.config.Browse
	}
	return Bigger.Route(browse, Map{
		"{code}": code, "name": name,  "token": token,
	})
}

func (base *defaultFileBase) Preview(code string, width,height,tttt int64, args Map, expires ...time.Duration) (string) {
	expiry := time.Hour*24

	if base.connect.config.Expiry != "" {
		if vv,err := Bigger.Timing(base.connect.config.Expiry); err == nil {
			expiry = vv
		}
	}
	if len(expires) > 0 {
		expiry = expires[0]
	}

	//ip, expiry, name
	deadline := time.Now().Add(expiry)
	tokens := []string{
		fmt.Sprintf("%v", args["ip"]),
		fmt.Sprintf("%d", deadline.UnixNano()),
	}
	data := Bigger.Decode(code)
	if data != nil {
		tokens = append(tokens, data.Name)
	}
	token := Bigger.Encrypts(tokens)

	preview := previewName
	if base.connect.config.Preview != "" {
		preview = base.connect.config.Preview
	}

	return Bigger.Route(preview, Map{
		"{code}": code, "{size}": []int64{ width, height, tttt },
		"token": token,
	})
}


func (base *defaultFileBase) storaging(data *FileCoding) (string,string,string,*Error) {
	if ring := base.connect.hashring.Locate(data.Full()); ring != "" {

		spath := path.Join(base.connect.setting.Storage, ring)
		sfile := path.Join(spath, data.Full())

		// //创建目录
		erro := os.MkdirAll(spath, 0777)
		if erro != nil {
			return "","","",Bigger.Erring("生成目录失败")
		}

		return ring,spath, sfile, nil
	}

	return "","","",Bigger.Erring("配置异常")
}

func (base *defaultFileBase) thumbnailing(data *FileCoding, width, height, tttt int64) (string,string,string,*Error) {
	if ring := base.connect.hashring.Locate(data.Full()); ring != "" {

		// data.Type = "jpg"	//不能直接改，因为是*data，所以扩展名不同的，生成缩图就有问题了，ring变了
		namenoext := strings.TrimSuffix(data.Name, "." + data.Type)

		tpath := path.Join(base.connect.setting.Thumbnail, ring, namenoext)
		tname := fmt.Sprintf("%d-%d-%d.%s", width, height, tttt, data.Type)
		tfile := path.Join(tpath, tname)

		// //创建目录
		erro := os.MkdirAll(tpath, 0777)
		if erro != nil {
			return "","","",Bigger.Erring("生成目录失败")
		}

		return ring,tpath, tfile, nil
	}

	return "","","",Bigger.Erring("配置异常")
}

//-------------------- defaultFileBase end -------------------------




