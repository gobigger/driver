package logger_default


import (
    . "github.com/gobigger/bigger"
    "github.com/gobigger/bigger/log"
    "strings"
)

//默认logger驱动

type (
    defaultLoggerDriver struct {}
    defaultLoggerConnect struct {
        config  LoggerConfig
        logger  *log.Logger
    }
)


func (driver *defaultLoggerDriver) Connect(config LoggerConfig) (LoggerConnect,*Error) {
    return &defaultLoggerConnect{
        config: config,
    },nil
}


//打开连接
func (connect *defaultLoggerConnect) Open() *Error {
    connect.logger = log.NewLogger()

    logConfig := &log.ConsoleConfig{
        Json: false, Format: "%time% [%type%] %body%",
    }

    if connect.config.Format != "" {
        logConfig.Format = connect.config.Format
    }

    level := log.LoggerLevel(connect.config.Level)
    connect.logger.Attach("console", level, logConfig)

    // connect.logger.SetAsync()

    return nil
}

func (connect *defaultLoggerConnect) Health() (*LoggerHealth,*Error) {
	// connect.mutex.RLock()
	// defer connect.mutex.RUnlock()
	return &LoggerHealth{ Workload: 0 },nil
}
//关闭连接
func (connect *defaultLoggerConnect) Close() *Error {
    connect.logger.Flush()
    return nil
}


func (connect *defaultLoggerConnect) Debug(body string, args ...Any) {
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
func (connect *defaultLoggerConnect) Trace(body string, args ...Any) {
    
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
func (connect *defaultLoggerConnect) Info(body string, args ...Any) {
    
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
func (connect *defaultLoggerConnect) Warning(body string, args ...Any) {
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
func (connect *defaultLoggerConnect) Error(body string, args ...Any) {
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



