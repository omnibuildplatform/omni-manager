package routers

import (
	"github.com/omnibuildplatform/omni-manager/controllers"
	"github.com/omnibuildplatform/omni-manager/docs"
	"github.com/omnibuildplatform/omni-manager/models"
	"github.com/omnibuildplatform/omni-manager/util"

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
	auth := r.Group(docs.SwaggerInfo.BasePath)
	{
		auth.GET("/v1/auth/loginok", controllers.AuthingLoginOk)
		auth.GET("/v1/auth/getDetail/:authingUserId", controllers.AuthingGetUserDetail)
		auth.Use(models.Authorize()) //
		auth.POST("/v1/auth/createUser", controllers.AuthingCreateUser)
	}
	//version 1 . call k8s api
	v1 := r.Group(docs.SwaggerInfo.BasePath + "/v1")
	{
		v1.Use(models.Authorize()) //
		v1.POST("/images/startBuild", controllers.StartBuild)
		v1.GET("/images/param/getBaseData/", controllers.GetBaseData)
		v1.GET("/images/param/getCustomePkgList/", controllers.GetCustomePkgList)
		v1.GET("/images/queryJobStatus/:name", controllers.QueryJobStatus)
		v1.GET("/images/queryJobLogs/:name", controllers.QueryJobLogs)
		v1.GET("/images/queryHistory/mine", controllers.QueryMyHistory)
	}
	//version 2. call owner api
	v2 := r.Group(docs.SwaggerInfo.BasePath + "/v2")
	{
		v2.Use(models.Authorize()) //
		v2.POST("/images/createJob", controllers.CreateJob)
		v2.GET("/images/getOne/:id", controllers.GetOne)
		v2.GET("/images/getJobParam/:id", controllers.GetJobParam)
		v2.GET("/images/getLogsOf/:id", controllers.GetJobLogs)
		v2.POST("/images/deleteJob", controllers.DeleteJobLogs)
		v2.GET("/images/getMySummary", controllers.GetMySummary)
		v2.DELETE("/images/stopJob/:id", controllers.StopJobBuild)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return r
}
