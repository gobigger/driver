package logger_file


import (
	"time"
    . "github.com/gobigger/bigger"
    "github.com/gobigger/bigger/log"
    "strings"
)

const (
    defaultErrorFile    = "store/logs/error.log"
    defaultWarningFile  = "store/logs/warning.log"
    defaultInfoFile     = "store/logs/info.log"
    defaultTraceFile    = "store/logs/trace.log"
    defaultDebugFile    = "store/logs/debug.log"
    defaultOutputFile   = "store/logs/output.log"
)

//默认logger驱动

type (
    fileLoggerDriver struct {}
    fileLoggerConnect struct {
        config  LoggerConfig
        logger  *log.Logger
    }
)


func (driver *fileLoggerDriver) Connect(config LoggerConfig) (LoggerConnect,*Error) {
    return &fileLoggerConnect{
        config: config,
    },nil
}


//打开连接
func (connect *fileLoggerConnect) Open() *Error {
    connect.logger = log.NewLogger()

    logFiles := map[int]string{}

    //错误日志
    if vv,ok := connect.config.Setting["error"].(string); ok {
        if vv != "" {
            logFiles[log.LEVEL_ERROR] = vv
        }
    } else if vv,ok := connect.config.Setting["error"].(bool); ok {
        if vv {
            logFiles[log.LEVEL_ERROR] = defaultErrorFile
        }
    } else {
        logFiles[log.LEVEL_ERROR] = defaultErrorFile
    }


    if vv,ok := connect.config.Setting["warning"].(string); ok {
        if vv != "" {
            logFiles[log.LEVEL_WARNING] = vv
        }
    } else if vv,ok := connect.config.Setting["warning"].(bool); ok {
        if vv {
            logFiles[log.LEVEL_WARNING] = defaultWarningFile
        }
    } else {
        logFiles[log.LEVEL_WARNING] = defaultWarningFile
    }


    
    if vv,ok := connect.config.Setting["info"].(string); ok {
        if vv != "" {
            logFiles[log.LEVEL_INFO] = vv
        }
    } else if vv,ok := connect.config.Setting["info"].(bool); ok {
        if vv {
            logFiles[log.LEVEL_INFO] = defaultInfoFile
        }
    } else {
        logFiles[log.LEVEL_INFO] = defaultInfoFile
    }


    if vv,ok := connect.config.Setting["trace"].(string); ok {
        if vv != "" {
            logFiles[log.LEVEL_TRACE] = vv
        }
    } else if vv,ok := connect.config.Setting["trace"].(bool); ok {
        if vv {
            logFiles[log.LEVEL_TRACE] = defaultTraceFile
        }
    } else {
        logFiles[log.LEVEL_TRACE] = defaultTraceFile
    }


    if vv,ok := connect.config.Setting["debug"].(string); ok {
        if vv != "" {
            logFiles[log.LEVEL_DEBUG] = vv
        }
    } else if vv,ok := connect.config.Setting["debug"].(bool); ok {
        if vv {
            logFiles[log.LEVEL_DEBUG] = defaultDebugFile
        }
    } else {
        logFiles[log.LEVEL_DEBUG] = defaultDebugFile
    }



	fileConfig := &log.FileConfig{
		// Filename : "logs/test.log",
		LevelFileName : logFiles,
		MaxSize : 1024*1024*100,
		MaxLine : 1000000,
		DateSlice : "day",
		Json: false, Format: "%time% [%type%] %body%",
    }


    if vv,ok := connect.config.Setting["output"].(string); ok {
        if vv != "" {
            fileConfig.Filename = vv
        }
    } else if vv,ok := connect.config.Setting["output"].(bool); ok {
        if vv {
            fileConfig.Filename = defaultOutputFile
        }
    }




    if connect.config.Format != "" {
        fileConfig.Format = connect.config.Format
    }
    
    //maxsize
    if vv,ok := connect.config.Setting["maxsize"].(string); ok && vv!="" {
        size := Bigger.Sizing(vv)
        if size > 0 {
            fileConfig.MaxSize = size
        }
    }
    //maxline
    if vv,ok := connect.config.Setting["maxline"].(int64); ok && vv > 0 {
        fileConfig.MaxLine = vv
    }
    if vv,ok := connect.config.Setting["segment"].(string); ok && vv!="" {
        fileConfig.DateSlice = log.CheckSlice(vv)
    }


    level := log.LoggerLevel(connect.config.Level)


	connect.logger.Attach("file", level, fileConfig)


    if connect.config.Console {
        //是否开启控制台日志
        consoleConfig := &log.ConsoleConfig{
            Json: false, Format: "%time% [%type%] %body%",
        }
        if connect.config.Format != "" {
            consoleConfig.Format = connect.config.Format
        }
        connect.logger.Attach("console", level, consoleConfig)
    }

    // connect.logger.SetAsync()
    
    return nil
}


func (connect *fileLoggerConnect) Health() (*LoggerHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &LoggerHealth{ Workload: 0 },nil
}

//关闭连接
func (connect *fileLoggerConnect) Close() *Error {
    //为了最后一条日志能正常输出
    time.Sleep(time.Microsecond*200)
    connect.logger.Flush()
    return nil
}







func (connect *fileLoggerConnect) Debug(body string, args ...Any) {
    if len(args)>0 && strings.Count(body, "%")==len(args) {
        connect.logger.Debugf(body, args...)
    } else {
        msgs := []string{ body }
        for _,v := range args {
            msgs = append(msgs, Bigger.ToString(v))
        }
        connect.logger.Debug(strings.Join(msgs, " "))
    }
}
func (connect *fileLoggerConnect) Trace(body string, args ...Any) {
    
    if len(args)>0 && strings.Count(body, "%")==len(args) {
        connect.logger.Tracef(body, args...)
    } else {
        msgs := []string{ body }
        for _,v := range args {
            msgs = append(msgs, Bigger.ToString(v))
        }
        connect.logger.Trace(strings.Join(msgs, " "))
    }
}
func (connect *fileLoggerConnect) Info(body string, args ...Any) {
    
    if len(args)>0 && strings.Count(body, "%")==len(args) {
        connect.logger.Infof(body, args...)
    } else {
        msgs := []string{ body }
        for _,v := range args {
            msgs = append(msgs, Bigger.ToString(v))
        }
        connect.logger.Info(strings.Join(msgs, " "))
    }
}
func (connect *fileLoggerConnect) Warning(body string, args ...Any) {
    if len(args)>0 && strings.Count(body, "%")==len(args) {
        connect.logger.Warningf(body, args...)
    } else {
        msgs := []string{ body }
        for _,v := range args {
            msgs = append(msgs, Bigger.ToString(v))
        }
        connect.logger.Warning(strings.Join(msgs, " "))
    }
}
func (connect *fileLoggerConnect) Error(body string, args ...Any) {
    if len(args)>0 && strings.Count(body, "%")==len(args) {
        connect.logger.Errorf(body, args...)
    } else {
        msgs := []string{ body }
        for _,v := range args {
            msgs = append(msgs, Bigger.ToString(v))
        }
        connect.logger.Error(strings.Join(msgs, " "))
    }
}



