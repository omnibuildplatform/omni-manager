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
		v1.POST("/meta/insert", controllers.Insert)
		v1.PUT("/meta/update", controllers.Update)
		v1.GET("/meta/get/:id", controllers.Read)
		v1.GET("/meta/query", controllers.Query)
		v1.DELETE("/meta/delete/:id", controllers.Delete)

	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
