package routers

import (
	"omni-manager/controllers"
	"omni-manager/docs"
	"omni-manager/util"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

//InitRouter init router
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(util.LoggerToFile())
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Title = util.GetConfig().AppName
	docs.SwaggerInfo.Description = util.GetConfig().AppName
	//version 1
	v1 := r.Group(docs.SwaggerInfo.BasePath)
	{
		v1.POST("/v1/images/startBuild", controllers.StartBuild)
		v1.GET("/v1/images/get/:id", controllers.Read)
		v1.GET("/v1/images/query", controllers.Query)
		v1.DELETE("/v1/images/delete/:id", controllers.Delete)
		v1.GET("/v1/images/param/getBaseData/", controllers.GetBaseData)
		v1.GET("/v1/images/param/getCustomePkgList/", controllers.GetCustomePkgList)
		v1.GET("/v1/images/queryJobStatus/:name", controllers.QueryJobStatus)
		v1.GET("/v1/images/queryJobLogs/:name", controllers.QueryJobLogs)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
