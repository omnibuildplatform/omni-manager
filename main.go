package main

import (
	"flag"
	"fmt"
	"omni-manager/image_monitor"
	"omni-manager/routers"
	"omni-manager/util"
)

func main() {
	var httpPort int
	flag.IntVar(&httpPort, "p", 0, "Input http port")
	flag.Parse()
	if httpPort == 0 {
		util.InitConfig()
		//use config port
		httpPort = util.GetConfig().AppPort
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
	//startup images status monitor
	go image_monitor.StartMonitor()

	//init router
	r := routers.InitRouter()
	util.Log.Infof(" startup meta service at port %s \n", address)
	if err := r.Run(address); err != nil {
		util.Log.Errorf("startup meta   service failed, err:%v\n", err)
	}

}
