package controllers

import (
	"fmt"
	"net/http"
	"omni-manager/models"
	"omni-manager/util"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary StartBuild Job
// @Description start a image build job
// @Tags  meta Manager
// @Param	body		body 	models.BuildParam	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v2/images/startBuild [post]
func StartBuildV2(c *gin.Context) {

	var imageInputData models.BuildParam
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	var insertData models.JobLog
	insertData.UserName = c.Keys["nm"].(string)
	insertData.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	insertData.Arch = imageInputData.Arch
	insertData.Release = imageInputData.Release
	insertData.BuildType = imageInputData.BuildType
	if len(insertData.Release) == 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "Release not allowed empty ", nil))
		return
	}

	if insertData.UserId <= 0 {
		c.JSON(http.StatusForbidden, util.ExportData(util.CodeStatusClientError, " forbidden ", nil))
		return
	}
	//check package validate
	validate := false
	for _, arch := range util.GetConfig().BuildParam.Arch {
		if arch == insertData.Arch {
			validate = true
			break
		}
	}
	if !validate {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "arch not supported  ", util.GetConfig().BuildParam.Arch))
		return
	}
	validate = false //reset for buildtype
	for _, buildtype := range util.GetConfig().BuildParam.BuildType {
		if buildtype == insertData.BuildType {
			validate = true
			break
		}
	}
	if !validate {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "buildType not supported  ", util.GetConfig().BuildParam.BuildType))
		return
	}
	insertData.BasePkg = strings.Join(util.GetConfig().DefaultPkgList.Packages, ",")
	insertData.CustomPkg = strings.Join(imageInputData.CustomPkg, ",")
	//-------- make custom rpms config first
	cm := models.MakeConfigMap(insertData.Release, imageInputData.CustomPkg)
	// //----------create job
	job, outPutname, err := models.MakeJob(cm, insertData.BuildType, insertData.Release)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "Create job Error", err))
		return
	}
	insertData.ConfigMapName = cm.Name
	insertData.UserName = c.Keys["nm"].(string)
	insertData.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	insertData.Status = models.JOB_STATUS_RUNNING
	insertData.JobName = job.GetName()
	insertData.CreateTime = job.GetCreationTimestamp().Time
	insertData.DownloadUrl = fmt.Sprintf(util.GetConfig().BuildParam.DownloadIsoUrl, insertData.Release, time.Now().Format("2006-01-02"), outPutname)
	err = models.AddJobLog(&insertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, job.GetName(), util.GetConfig().WSConfig))
}
