package main

import (
	"fmt"
	"omni-manager/models"
	"omni-manager/routers"
	"omni-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	util.InitConfig("")
}
func main() {

	if util.GetConfig().AppModel == "dev" {
		util.Log.SetLevel(logrus.DebugLevel)
		util.GetConfig().AppModel = gin.DebugMode
	} else {
		util.Log.SetLevel(logrus.WarnLevel)
		util.GetConfig().AppModel = gin.ReleaseMode
	}

	//init database
	err := util.InitDB()
	if err != nil {
		util.Log.Errorf("database connect failed , err:%v\n", err)
		return
	}

	err = models.CreateTables()
	if err != nil {
		util.Log.Errorf("database create tables failed , err:%v\n", err)
		return
	}
	//init redis
	err = util.InitRedis()
	if err != nil {
		util.Log.Errorf("Redis connect failed , err:%v\n", err)
		return
	}
	fmt.Println("===", util.GetConfig().AppModel)
	if util.GetConfig().AppModel != gin.DebugMode {
		//init customPkgs
		models.InitCustomPkgs()
	}
	//init Authing.cn config
	models.InitAuthing("", "")
	//init kubernetes client-go
	models.InitK8sClient()
	util.InitStatisticsLog()
	//startup a webscoket server to wait client ws
	go models.StartWebSocket()
	gin.SetMode(util.GetConfig().AppModel)
	r := routers.InitRouter()
	address := fmt.Sprintf(":%d", util.GetConfig().AppPort)
	util.Log.Printf(" startup meta http service at port %s .and %s mode \n", address, util.GetConfig().AppModel)
	if err := r.Run(address); err != nil {
		util.Log.Printf("startup meta  http service failed, err:%v\n", err)
	}
}
