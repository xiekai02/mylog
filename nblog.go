package nblog

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	LOGLEVELNONE int = iota
	LOGLEVELINFO
	LOGLEVELDEBUG
	LOGLEVELWARN
	LOGLEVELERROR
)

type Log struct {
	LogLevel int
	//LogSize    int64
	LogTime    int64
	LogTimeLen int64
	LogSizeLen int64
	LogFile    string
	File       *os.File
	LogMutex   sync.Mutex
}

func (log *Log) Open() bool {
	var err error
	log.File, err = os.OpenFile(log.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("open file failed:", err.Error())
		return false
	}
	info, err := log.File.Stat()
	if err != nil {
		return false
	}
	log.LogTime = info.Sys().(*syscall.Stat_t).Ctim.Sec

	return true
}
func (log *Log) Close() {
	log.File.Close()
}
func (log *Log) SetLogFile(file string) {
	log.LogFile = file
}
func (log *Log) SetTimeLen(len int64) {
	log.LogTimeLen = len
}
func (log *Log) SetSizeLen(len int64) {
	log.LogSizeLen = len
}
func (log *Log) SetLogLevel(level string) {
	switch level {
	case "INFO":
		log.LogLevel = LOGLEVELINFO
	case "DEBUG":
		log.LogLevel = LOGLEVELDEBUG
	case "WARN":
		log.LogLevel = LOGLEVELWARN
	case "ERROR":
		log.LogLevel = LOGLEVELERROR
	default:
		log.LogLevel = LOGLEVELERROR
	}
}
func (log *Log) GetFileSize() int64 {
	info, err := log.File.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
func (log *Log) Reopen() bool {
	log.File.Close()
	err := os.Remove(log.LogFile)
	if err != nil {
		return false
	}
	return log.Open()
}
func (log *Log) GetLevel(level int) string {
	switch level {
	case LOGLEVELINFO:
		return "INFO"
	case LOGLEVELDEBUG:
		return "DEBUG"
	case LOGLEVELWARN:
		return "WARN"
	case LOGLEVELERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
func (log *Log) Write(level int, s string) bool {
	if level < log.LogLevel {
		return true
	}
	log.LogMutex.Lock()
	defer log.LogMutex.Unlock()

	size := log.GetFileSize()
	curt := time.Now().Unix()
	if size > log.LogSizeLen || curt-log.LogTime > log.LogTimeLen {
		log.Reopen()
	}
	time := time.Now().Format("2006-01-02 15:04:05")
	_, fun, line := GetCallerInfo(3)
	data := fmt.Sprintf("[%s - %s %s:%d] %s\n", time, log.GetLevel(level), fun, line, s)
	_, err := log.File.WriteString(data)
	if err != nil {
		return false
	}
	log.File.Sync()
	return true
}

func (log *Log) LogInfo(format string, a ...any) bool {
	data := fmt.Sprintf(format, a...)
	return log.Write(LOGLEVELINFO, data)
}
func (log *Log) LogDebug(format string, a ...any) bool {
	data := fmt.Sprintf(format, a...)
	return log.Write(LOGLEVELDEBUG, data)
}
func (log *Log) LogWarn(format string, a ...any) bool {
	data := fmt.Sprintf(format, a...)
	return log.Write(LOGLEVELWARN, data)
}
func (log *Log) LogError(format string, a ...any) bool {
	data := fmt.Sprintf(format, a...)
	return log.Write(LOGLEVELERROR, data)
}
func GetCallerInfo(level int) (string, string, int) {
	pc, file, line, _ := runtime.Caller(level)
	funcName := runtime.FuncForPC(pc).Name()
	return file, funcName, line
}
