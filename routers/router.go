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
	r := gin.Default()
	r.Use(util.LoggerToFile())
	docs.SwaggerInfo.BasePath = "/api/v1"
	//version 1
	v1 := r.Group(docs.SwaggerInfo.BasePath)
	{
		v1.POST("/images/startBuild", controllers.StartBuild)
		v1.PUT("/images/update", controllers.Update)
		v1.GET("/images/get/:id", controllers.Read)
		v1.GET("/images/query", controllers.Query)
		v1.DELETE("/images/delete/:id", controllers.Delete)
		v1.GET("/images/param/getBaseData/", controllers.GetBaseData)
		v1.GET("/images/param/getCustomePkgList/", controllers.GetCustomePkgList)
		v1.GET("/images/queryJobStatus/:name", controllers.QueryJobStatus)

	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
