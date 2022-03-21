package main

import (
	"flag"
	"fmt"
	"omni-manager/models"
	"omni-manager/routers"
	"omni-manager/util"

	"github.com/gin-gonic/gin"
)

func main() {
	var httpPort int
	flag.IntVar(&httpPort, "p", 0, "Input http port")
	flag.Parse()
	//load config file
	util.InitConfig()
	if httpPort <= 0 {
		//use flag port first ,if not then use config port
		httpPort = util.GetConfig().AppPort
	}
	if httpPort <= 0 {
		//if config port not set,then set a default 8080
		httpPort = 8080
	}
	address := fmt.Sprintf(":%d", httpPort)
	//init database
	err := util.InitDB()
	if err != nil {
		util.Log.Errorf("database connect failed , err:%v\n", err)
		return
	}
	//init redis
	err = util.InitRedis()
	if err != nil {
		util.Log.Errorf("Redis connect failed , err:%v\n", err)
		return
	}
	//init kubernetes client-go
	models.InitK8sClient()
	//startup a webscoket server to wait client ws
	go models.StartWebSocket()
	//init router
	gin.SetMode(util.GetConfig().AppModel)
	r := routers.InitRouter()
	util.Log.Errorf(" startup meta http service at port %s .and %s mode \n", address, util.GetConfig().AppModel)
	if err := r.Run(address); err != nil {
		util.Log.Fatalf("startup meta  http service failed, err:%v\n", err)
	}
}
