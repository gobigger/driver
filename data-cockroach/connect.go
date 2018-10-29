package data_cockroach

import (
	"sync"
	. "github.com/gobigger/bigger"
	"database/sql"
)

type (
	//数据库连接
	CockroachConnect struct {
		mutex		sync.RWMutex
		actives	int64

		name		string
		config		DataConfig
		schema		string

		//数据库对象
		db  	*sql.DB
	}
)

//打开连接
func (connect *CockroachConnect) Open() *Error {
	db, err := sql.Open(SQLDRIVER, connect.config.Url)
	if err != nil {
		return Bigger.Erred(err)
	} else {
		connect.db = db
		return nil
	}
}
//健康检查
func (connect *CockroachConnect) Health() (*DataHealth,*Error) {
	connect.mutex.RLock()
	defer connect.mutex.RUnlock()
	return &DataHealth{ Workload: connect.actives },nil
}

//关闭连接
func (connect *CockroachConnect) Close() *Error {
	if connect.db != nil {
		err := connect.db.Close()
		if err != nil {
			return Bigger.Erred(err)
		}
		connect.db = nil
	}
	return nil
}


func (connect *CockroachConnect) Base(cache CacheBase) (DataBase) {
	connect.mutex.Lock()
	connect.actives++
	connect.mutex.Unlock()

	return &CockroachBase{connect, connect.name, connect.schema, nil, cache, false, []DataTrigger{}, nil}
}

