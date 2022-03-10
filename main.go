package main

import (
	"flag"
	"fmt"
	"omni-manager/routers"
	"omni-manager/util"
)

func main() {
	var httpPort int
	flag.IntVar(&httpPort, "p", 0, "Input http port")
	flag.Parse()
	if httpPort == 0 {
		//use config port
		httpPort = util.GetConfig().AppPort
	}
	address := fmt.Sprintf(":%d", httpPort)

	//init database
	err := util.InitDB()
	if err != nil {
		util.Log.Errorf("database startup failed , err:%v\n", err)
		return
	}
	//init router
	r := routers.InitRouter()
	util.Log.Infof(" startup meta service at port %s \n", address)
	if err := r.Run(address); err != nil {
		util.Log.Errorf("startup meta   service failed, err:%v\n", err)
	}

}
