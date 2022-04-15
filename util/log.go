package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

type StatisticsData struct {
	UserId        int
	UserName      string
	UserProvider  string
	UserEmail     string
	OperationTime string
	EventType     string
	State         string
	StateMessage  string
	Body          interface{}
}

//statistics log
var SLog *logrus.Logger

func init() {
	initLogger()
}

func initLogger() {
	logFilePath := ""
	if dir, err := os.Getwd(); err == nil {
		logFilePath = dir + "/logs/"
	}
	if err := os.MkdirAll(logFilePath, 0777); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	logFileName := time.Now().Format("2006-01-02") + ".log"
	//log file
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			fmt.Println(err.Error())
		}
	}
	//open file
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}

	//new log
	Log = logrus.New()
	Log.Out = io.MultiWriter(os.Stdout, src)
	Log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return
}
func InitStatisticsLog() {
	//-=----------------------------------------

	if err := os.MkdirAll(GetConfig().Statistic.Dir, 0777); err != nil {
		Log.Errorf("InitStatisticsLog Error %v", err)
		os.Exit(1)
	}
	logFileName := GetConfig().AppName + "-" + time.Now().Format("2006-01-02") + ".log"
	//log file
	fileName := path.Join(GetConfig().Statistic.Dir, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			fmt.Println(err.Error())
		}
	}
	//open file
	src, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("err", err)
	}
	//new log
	SLog = logrus.New()
	SLog.Out = io.MultiWriter(os.Stdout, src)
	SLog.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: true,
		PrettyPrint:      true,
	})
}
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		//日志格式
		Log.Infof("| %3d | %13v | %15s | %s | %s |",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)

	}
}

func StatisticsLog(sd *StatisticsData) error {
	if sd.State == "" {
		sd.State = "success"
	}
	mapData := make(map[string]interface{})
	mapData["OperationTime"] = fmt.Sprintf("%v", sd.OperationTime)
	mapData["UserId"] = fmt.Sprintf("%v", sd.UserId)
	mapData["UserName"] = fmt.Sprintf("%v", sd.UserName)
	mapData["UserProvider"] = fmt.Sprintf("%v", sd.UserProvider)
	mapData["EventType"] = fmt.Sprintf("%v", sd.EventType)
	mapData["Body"] = sd.Body
	mapData["AppId"] = GetConfig().AuthingConfig.AppID
	mapData["State"] = sd.State
	mapData["StateMessage"] = sd.StateMessage
	data, err := json.Marshal(mapData)
	if err != nil {
		Log.Error("StatisticsLog Marshal err: ", err)
		return err
	}
	SLog.Out.Write(data)
	SLog.Out.Write([]byte("\n"))
	return nil
}
