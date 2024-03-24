package util

//使用logrus进行日志处理

import (
	"io"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

var logFilePath = "log"
var logFileName = "experiment.log"

var Logger = logrus.New()

func init() {
	// 获取当前路径
	dir, _ := os.Getwd()
	logFilePath = path.Join(dir, logFilePath)
	fileName := path.Join(logFilePath, logFileName)
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}
	multiWriter := io.MultiWriter(src, os.Stdout)

	Logger.Out = multiWriter
	Logger.SetLevel(logrus.DebugLevel)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
