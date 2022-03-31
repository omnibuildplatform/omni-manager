package routers

import (
	"omni-manager/controllers"
	"omni-manager/docs"
	"omni-manager/models"
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
	docs.SwaggerInfo.Description = "set token name: 'Authorization' at header "
	//version 1
	v1 := r.Group(docs.SwaggerInfo.BasePath)
	{
		v1.Use(models.Authorize()) //
		v1.POST("/v1/images/startBuild", controllers.StartBuild)
		v1.GET("/v1/images/get/:id", controllers.Read)
		v1.GET("/v1/images/query", controllers.Query)
		v1.GET("/v1/images/param/getBaseData/", controllers.GetBaseData)
		v1.GET("/v1/images/param/getCustomePkgList/", controllers.GetCustomePkgList)
		v1.GET("/v1/images/queryJobStatus/:name", controllers.QueryJobStatus)
		v1.GET("/v1/images/queryJobLogs/:name", controllers.QueryJobLogs)
		v1.GET("/v1/images/queryHistory/mine", controllers.QueryMyHistory)

	}
	auth := r.Group(docs.SwaggerInfo.BasePath)
	{
		auth.GET("/v1/auth/loginok", controllers.AuthingLoginOk)
		auth.GET("/v1/auth/getDetail/:authingUserId", controllers.AuthingGetUserDetail)
		auth.Use(models.Authorize()) //
		auth.POST("/v1/auth/createUser", controllers.AuthingCreateUser)

	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
