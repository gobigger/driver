package data_postgres


import (
	. "github.com/gobigger/bigger"
	"strings"
)

type (
	PostgresDriver struct {
	}
)

//驱动连接
func (drv *PostgresDriver) Connect(name string, config DataConfig) (DataConnect, *Error) {

	//支持自定义的schema，相当于数据库名
	schema := SCHEMA

	for _,s := range schemas {
		if strings.HasPrefix(config.Url, s) {
			config.Url = strings.Replace(config.Url, s, "postgres://", 1)
		}
	}

	if vv,ok := config.Setting["schema"].(string); ok && vv!="" {
		schema = vv
	}

	// if config.Url != "" {
	// 	durl,err := url.Parse(config.Url)
	// 	if err == nil {
	// 		if len(durl.Path) >= 1 {
	// 			schema = durl.Path[1:]
	// 		}
	// 	}
	// } else if vv,ok := config.Setting["schema"].(string); ok && vv != "" {
	// 	schema = vv
	// }

	return &PostgresConnect{
		name: name, config: config, schema: schema, db: nil, actives: int64(0),
	}, nil
}



