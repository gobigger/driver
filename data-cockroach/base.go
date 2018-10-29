package data_cockroach


import (
	. "github.com/gobigger/bigger"
	"database/sql"
	"fmt"
	"strings"
	"encoding/json"
	"strconv"
	"time"
)

type (
	CockroachBase struct {
		connect	*CockroachConnect

		name    string
		schema	string
		
		tx      *sql.Tx
		cache   CacheBase

		//是否手动提交事务，否则为自动
		//当调用begin时， 自动变成手动提交事务
		//triggers保存待提交的触发器，手动下有效
		manual      bool
		triggers    []DataTrigger

		lastError	*Error
	}
)




//记录触发器
func (base *CockroachBase) trigger(name string, values ...Map) {
	value := Map{}
	if len(values) > 0 {
		value = values[0]
	}
	base.triggers = append(base.triggers, DataTrigger{ Name: name, Value: value })
}


//查询表，支持多个KEY遍历
func (base *CockroachBase) table(name string) (Map) {
	keys := []string{
		fmt.Sprintf("%s.%s", base.name, name),
		fmt.Sprintf("*.%s", name),
		name,
	}

	for _,key := range keys {
		if cfg := Bigger.Table(key); cfg != nil {
			return cfg
		}
	}

	return nil
}
func (base *CockroachBase) view(name string) (Map) {
	keys := []string{
		fmt.Sprintf("%s.%s", base.name, name),
		fmt.Sprintf("*.%s", name),
		name,
	}

	for _,key := range keys {
		if cfg := Bigger.View(key); cfg != nil {
			return cfg
		}
	}

	return nil
}
func (base *CockroachBase) model(name string) (Map) {
	keys := []string{
		fmt.Sprintf("%s.%s", base.name, name),
		fmt.Sprintf("*.%s", name),
		name,
	}

	for _,key := range keys {
		if cfg := Bigger.Model(key); cfg != nil {
			return cfg
		}
	}

	return nil
}


func (base *CockroachBase) error(key string, err error, args ...Any) {
	if err != nil {
		//出错自动取消事务
		base.Cancel()

		errors := []Any{ err }
		errors = append(errors, args...)

		base.lastError = Bigger.Erred(err)
		Bigger.Warning(key, errors...)
	}
}








//关闭数据库
func (base *CockroachBase) Close() *Error {
	base.connect.mutex.Lock()
	base.connect.actives--
	base.connect.mutex.Unlock()

	//好像目前不需要关闭什么东西
	if base.tx != nil {
		//关闭时候,一定要提交一次事务
		//如果手动提交了, 这里会失败, 问题不大
		//如果没有提交的话, 连接不会交回连接池. 会一直占用
		base.Cancel()
	}

	if base.cache != nil {
		base.cache.Close()
	}

	return nil
}
func (base *CockroachBase) Erred() (*Error) {
	err := base.lastError
	base.lastError = nil
    return err
}


//ID生成器
func (base *CockroachBase) Serial(key string) (int64) {

	exec,err := base.begin()
	if err != nil {
		base.error("data.serial.begin", err, key)
		return 0
	}


	serial := "serial"
	if base.connect.config.Serial != "" {
		serial = base.connect.config.Serial
	} else if vv,ok := base.connect.config.Setting["serial"].(string); ok && vv != "" {
		serial = vv
	}

	sql := fmt.Sprintf(
		`INSERT INTO %v(key,seq) VALUES ($1,$2) ON CONFLICT (key) DO UPDATE SET seq=%v.seq+excluded.seq RETURNING seq;`,
			serial, serial,
		)
	args := []Any{ key, int64(1) }
	row := exec.QueryRow(sql, args...)

	seq := int64(0)

	err = row.Scan(&seq)
	if err != nil {
		base.error("data.serial.scan", err, key)
		return 0
	}

	return seq
}




//获取表对象
func (base *CockroachBase) Table(name string) (DataTable) {
	if config := base.table(name); config != nil {
		//模式，表名
		schema, table, key, fields := base.schema, name, "id", Map{}
		if n,ok := config["schema"].(string); ok {
			schema = n
		}
		if n,ok := config["table"].(string); ok {
			table = n
		}
		if n,ok := config["key"].(string); ok {
			key = n
		}
		if n,ok := config["fields"].(Map); ok {
			fields = n
		}

		return &CockroachTable{
			CockroachView{base, name, schema, table, key, fields},
		}
	} else {
		panic("[数据]表不存在")
	}
}

//获取模型对象
func (base *CockroachBase) View(name string) (DataView) {
	if config := base.view(name); config != nil {

		//模式，表名
		schema, view, key, fields := base.schema, name, "id", Map{}
		if n,ok := config["schema"].(string); ok {
			schema = n
		}
		if n,ok := config["view"].(string); ok {
			view = n
		}
		if n,ok := config["key"].(string); ok {
			key = n
		}
		if n,ok := config["fields"].(Map); ok {
			fields = n
		}

		return &CockroachView{
			base, name, schema, view, key, fields,
		}
	} else {
		panic("[数据]视图不存在")
	}
}

//获取模型对象
func (base *CockroachBase) Model(name string) (DataModel) {
	if config := base.model(name); config != nil {

		//模式，表名
		schema, model, key, fields := base.schema, name, "id", Map{}
		if n,ok := config["schema"].(string); ok {
			schema = n
		}
		if n,ok := config["model"].(string); ok {
			model = n
		}
		if n,ok := config["key"].(string); ok {
			key = n
		}
		if n,ok := config["field"].(Map); ok {
			fields = n
		}

		return &CockroachModel{
			base, name, schema, model, key, fields,
		}
	} else {
		panic("[数据]模型不存在")
	}
}

//是否开启缓存
// func (base *CockroachBase) Cache(use bool) (DataBase) {
// 	base.caching = use
// 	return base
// }



//开启手动模式
func (base *CockroachBase) Begin() (*sql.Tx, *Error) {
	base.lastError = nil

	if _,err := base.begin(); err != nil {
		return nil, Bigger.Erred(err)
	}
	
	base.manual = true
	return base.tx, nil
}




//注意，此方法为实际开始事务
func (base *CockroachBase) begin() (CockroachExecutor, error) {
	if base.manual {
		if base.tx == nil {
			tx,err := base.connect.db.Begin()
			if err != nil {
				return nil, err
			}
			base.tx = tx
		}
		return base.tx, nil
	} else {
		return base.connect.db, nil
	}
}




//提交事务
func (base *CockroachBase) Submit() (*Error) {
	defer func() {
		//不管成功失败，都清事务
		base.tx = nil
		base.manual = false
	}()

	if base.tx == nil {
		return Bigger.Erring("[数据]无效事务")
	}

	err := base.tx.Commit()
	if err != nil {
		return Bigger.Erred(err)
	}

	//提交事务后,要把触发器都发掉
	for _,trigger := range base.triggers {
		Bigger.Trigger(trigger.Name, trigger.Value)
	}
	//清空触发器
	base.triggers = []DataTrigger{}

	return nil
}


//取消事务
func (base *CockroachBase) Cancel() (*Error) {

	if base.tx == nil {
		return Bigger.Erring("[数据]无效事务")
	}

	err := base.tx.Rollback()
	if err != nil {
		return Bigger.Erred(err)
	}

	//提交后,要清掉事务
	base.tx = nil
	base.manual = false

	return nil
}


















//创建的时候,也需要对值来处理,
//数组要转成{a,b,c}格式,要不然不支持
//json可能要转成字串才支持
func (base *CockroachBase) packing(value Map) (Map) {

	newValue := Map{}

	for k,v := range value {
		switch t := v.(type) {
		case []string: {
			newValue[k] = fmt.Sprintf(`{%s}`, strings.Join(t, `,`))
		}
		case []bool: {
			arr := []string{}
			for _,v := range t {
				if v {
					arr = append(arr, "TRUE")
				} else {
					arr = append(arr, "FALSE")
				}
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []int: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, strconv.Itoa(v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []int8: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, fmt.Sprintf("%v", v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []int16: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, fmt.Sprintf("%v", v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []int32: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, fmt.Sprintf("%v", v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []int64: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, fmt.Sprintf("%v", v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []float32: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, fmt.Sprintf("%v", v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case []float64: {
			arr := []string{}
			for _,v := range t {
				arr = append(arr, fmt.Sprintf("%v", v))
			}

			newValue[k] = fmt.Sprintf("{%s}", strings.Join(arr, ","))
		}
		case Map: {
			b,e := json.Marshal(t)
			if e == nil {
				newValue[k] = string(b)
			} else {
				newValue[k] = "{}"
			}
		}
		case []Map: {
			b,e := json.Marshal(t)
			if e == nil {
				newValue[k] = string(b)
			} else {
				newValue[k] = "[]"
			}
		}
		default:
			newValue[k] = t
		}
	}
	return newValue
}



//楼上写入前要打包处理值
//这里当然 读取后也要解包处理
func (base *CockroachBase) unpacking(keys []string, vals []interface{}) (Map) {

	m := Map{}

	for i,n := range keys {
		switch v := vals[i].(type) {
		case time.Time:
			m[n] = v.Local()	//设置为本地时间，因为cockroa目前存的时间，全部是utc时间
			//m[n] = v
		case string: {
			m[n] = v
		}
		case []byte: {
			m[n] = string(v)
		}
		default:
			m[n] = v
		}
	}

	return m
}









//把MAP编译成sql查询条件
func (base *CockroachBase) parsing(i int,args ...Any) (string,[]interface{},string,error) {
	sql,val,odr,err := Bigger.Query(args...)
	if err != nil {
		return "",nil,"",err
	}

	//结果要处理一下，字段包裹、参数处理
	sql = strings.Replace(sql, DELIMS, `"`, -1)
	odr = strings.Replace(odr, DELIMS, `"`, -1)
	for range val {
		sql = strings.Replace(sql, "?", fmt.Sprintf("$%d", i), 1)
		i++
	}

	return sql,val,odr,nil
}


















// //获取relate定义的parents
// func (base *CockroachBase) parents(name string) (Map) {
// 	values := Map{}

// 	if config,ok := base.tables(name); ok {
// 		if fields,ok := config["fields"].(Map); ok {
// 			base.parent(name, fields, []string{}, values)
// 		}
// 	}

// 	return values;
// }

// //获取relate定义的parents
// func (base *CockroachBase) parent(table string, fields Map, tree []string, values Map) {
// 	for k,v := range fields {
// 		config := v.(Map)
// 		trees := append(tree, k)

// 		if config["relate"] != nil {

// 			relates := []Map{}

// 			switch ttts := config["relate"].(type) {
// 			case Map:
// 				relates = append(relates, ttts)
// 			case []Map:
// 				for _,ttt := range ttts {
// 					relates = append(relates, ttt)
// 				}
// 			}

// 			for i,relating := range relates {

// 				//relating := config["relate"].(Map)
// 				parent := relating["parent"].(string)

// 				//要从模型定义中,把所有父表的 schema, table 要拿过来
// 				if tableConfig,ok := base.tables(parent); ok {

// 					schema,table := SCHEMA,parent
// 					if tableConfig["schema"] != nil && tableConfig["schema"] != "" {
// 						schema = tableConfig["schema"].(string)
// 					}
// 					if tableConfig["table"] != nil && tableConfig["table"] != "" {
// 						table = tableConfig["table"].(string)
// 					}

// 					//加入列表，带上i是可能有多个字段引用同一个表？还是引用多个表？
// 					values[fmt.Sprintf("%v:%v", strings.Join(trees, "."), i)] = Map{
// 						"schema": schema, "table": table,
// 						"field": strings.Join(trees, "."), "relate": relating,
// 					}
// 				}
// 			}


// 		} else {
// 			if json,ok := config["json"].(Map); ok {
// 				base.parent(table, json, trees, values)
// 			}
// 		}
// 	}
// }


// //获取relate定义的childs
// func (base *CockroachBase) childs(model string) (Map) {
// 	values := Map{}

// 	for modelName,modelConfig := range base.bonder.tables {

// 		schema,table := SCHEMA,modelName
// 		if modelConfig["schema"] != nil && modelConfig["schema"] != "" {
// 			schema = modelConfig["schema"].(string)
// 		}
// 		if modelConfig["table"] != nil && modelConfig["table"] != "" {
// 			table = modelConfig["table"].(string)
// 		}

// 		if fields,ok := modelConfig["field"].(Map); ok {
// 			base.child(model, modelName, schema, table, fields, []string{ }, values)
// 		}
// 	}

// 	return values;
// }

// //获取relate定义的child
// func (base *CockroachBase) child(parent,model,schema,table string, configs Map, tree []string, values Map) {
// 	for k,v := range configs {
// 		config := v.(Map)
// 		trees := append(tree, k)

// 		if config["relate"] != nil {


// 			relates := []Map{}

// 			switch ttts := config["relate"].(type) {
// 			case Map:
// 				relates = append(relates, ttts)
// 			case []Map:
// 				for _,ttt := range ttts {
// 					relates = append(relates, ttt)
// 				}
// 			}

// 			for i,relating := range relates {

// 				//relating := config["relate"].(Map)

// 				if relating["parent"] == parent {
// 					values[fmt.Sprintf("%v:%v:%v", model, strings.Join(trees, "."), i)] = Map{
// 						"schema": schema, "table": table,
// 						"field": strings.Join(trees, "."), "relate": relating,
// 					}
// 				}
// 			}

// 		} else {
// 			if json,ok := config["json"].(Map); ok {
// 				base.child(parent,model,schema,table,json, trees, values)
// 			}
// 		}
// 	}
// }


