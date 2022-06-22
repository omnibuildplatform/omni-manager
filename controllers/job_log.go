package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/omnibuildplatform/omni-manager/models"
	"github.com/omnibuildplatform/omni-manager/util"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// @Summary StartBuild Job
// @Description start a image build job
// @Tags  v1 version
// @Param	body		body 	models.BuildParam	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v1/images/startBuild [post]
func StartBuild(c *gin.Context) {

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
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "arch not be supported  ", util.GetConfig().BuildParam.Arch))
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
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "buildType not be supported  ", util.GetConfig().BuildParam.BuildType))
		return
	}
	insertData.BasePkg = strings.Join(util.GetConfig().DefaultPkgList.Packages, ",")
	insertData.CustomPkg = strings.Join(imageInputData.CustomPkg, ",")
	//-------- make custom rpms config first
	cm := models.MakeConfigMap(insertData.Release, imageInputData.CustomPkg)
	// //----------create job
	job, err := models.MakeJob(cm, insertData.BuildType, insertData.Release)
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
	insertData.DownloadUrl = "/api/v3/getRepositoryDownlad/" + insertData.JobName
	err = models.AddJobLog(&insertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}
	sd := util.StatisticsData{}
	sd.UserName = insertData.UserName
	sd.UserId = insertData.UserId
	sd.EventType = "使用v1构建"
	sd.Body = fmt.Sprintf("jobID:%s", job.Name)
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, job.GetName(), util.GetConfig().WSConfig))
}

// @Summary QueryJobStatus
// @Description QueryJobStatus for given job name
// @Tags  v1 version
// @Param	name		path 	string	true		"The name for job"
// @Param	id		query 	string	false		"The id for job in database. "
// @Param	ns		query 	string	false		"job namespace "
// @Accept json
// @Produce json
// @Router /v1/images/queryJobStatus/{name} [get]
func QueryJobStatus(c *gin.Context) {
	jobname := c.Param("name")
	if jobname == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job name must be fill:", nil))
		return
	}
	jobNamespace, _ := c.GetQuery("ns")
	if jobNamespace == "" {
		//if no special,then use config namespace
		jobNamespace = util.GetConfig().K8sConfig.Namespace
	}
	result, job, err := models.CheckPodStatus(jobNamespace, jobname)

	// buildLog, err := models.GetJobLogByJobName(jobname)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " not found job name:", jobname))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result, job))
}

// @Summary QueryJobLogs
// @Description QueryJobLogs for given job name
// @Tags  v1 version
// @Param	name		path 	string	true		"The name for job"
// @Accept json
// @Produce json
// @Router /v1/images/queryJobLogs/{name} [get]
func QueryJobLogs(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job name must be fill:", nil))
		return
	}
	listopt := metav1.ListOptions{}
	listopt.LabelSelector = "job-name=" + name
	pods, err := models.GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).List(context.TODO(), listopt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "", err))
		return
	}
	buf := new(bytes.Buffer)
	for _, pod := range pods.Items {
		buf.WriteString(fmt.Sprintf("------------------- pod.name:%s \n---------", pod.Name))
		req := models.GetClientSet().CoreV1().Pods(util.GetConfig().K8sConfig.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(context.TODO())
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "", err))
			return
		}
		defer podLogs.Close()
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "", err))
			return
		}
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", buf.String()))
}

// @Summary GetBaseData param
// @Description get architecture, release Version, output Format ,and default package name list
// @Tags  v1 version
// @Accept json
// @Produce json
// @Router /v1/images/param/getBaseData/ [get]
func GetBaseData(c *gin.Context) {

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal,
		"ok", util.GetConfig().BuildParam, util.GetConfig().DefaultPkgList, models.CustomSigList))
}

// @Summary GetCustomePkgList param
// @Description get custom package name list
// @Tags  v1 version
// @Param	arch		query 	string	true		" arch ,e g:x86_64"
// @Param	release		query 	string	true		"release  "
// @Param	sig		query 	string	true		"custom group  "
// @Accept json
// @Produce json
// @Router /v1/images/param/getCustomePkgList/ [get]
func GetCustomePkgList(c *gin.Context) {

	arch := c.Query("arch")
	release := c.Query("release")
	sig := c.Query("sig")

	customlist, err := models.GetCustomePkgList(release, arch, sig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "查询CustmPkg"
	sd.Body = fmt.Sprintf("release: %s, arch:%s, sig:%s", release, arch, sig)
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	util.StatisticsLog(&sd)

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", customlist))
}

// @Summary QueryMyHistory
// @Description Query My History
// @Tags  v1 version
// @Param	arch		query 	string	false		"arch"
// @Param	status		query 	string	false		"status"
// @Param	type		query 	string	false		"build type"
// @Param	nameordesc		query 	string	false		"name or desc"
// @Param	offset		query 	int	false		"offset "
// @Param	limit		query 	int	false		"limit"
// @Accept json
// @Produce json
// @Router /v1/images/queryHistory/mine [get]
func QueryMyHistory(c *gin.Context) {
	//...... emplty . wait for query param
	var UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	if UserId <= 0 {
		c.JSON(http.StatusForbidden, util.ExportData(util.CodeStatusClientError, " forbidden ", nil))
		return
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}

	nameordesc := c.Query("nameordesc")
	queryJobLog := new(models.JobLog)
	queryJobLog.UserId = UserId
	queryJobLog.Arch = c.Query("arch")
	queryJobLog.Status = c.Query("status")
	queryJobLog.BuildType = c.Query("type")
	total, result, err := models.GetMyJobLogs(queryJobLog, nameordesc, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}

	sd := util.StatisticsData{}
	sd.UserId = UserId
	sd.EventType = "查询自己的构建历史"
	sd.Body = fmt.Sprintf("offset: %d, limit:%d, result number:%d", offset, limit, len(result))
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	util.StatisticsLog(&sd)
	if len(result) == 0 {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", []interface{}{}, 0))
	} else {
		for _, item := range result {
			item.DownloadUrl = util.GetConfig().BuildServer.OmniRepoAPI + "/images/query?externalID=" + item.JobName
		}
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result, total))
	}

}
