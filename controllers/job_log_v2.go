package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"omni-manager/models"
	"omni-manager/util"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary Create Job
// @Description start a image build job
// @Tags  v2 job
// @Param	body		body 	models.BuildParam	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v2/images/createJob [post]
func CreateJob(c *gin.Context) {

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
	insertData.JobLabel = imageInputData.Label
	insertData.JobDesc = imageInputData.Desc
	if insertData.JobLabel == "" {
		insertData.JobLabel = insertData.UserName + "_" + insertData.Arch + "_" + insertData.Release
	}
	if insertData.JobDesc == "" {
		insertData.JobDesc = "this image was built by Omni Build Platform"
	}
	insertData.CreateTime = time.Now()
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
	specMap := make(map[string]interface{})
	specMap["version"] = insertData.Release
	specMap["packages"] = append(imageInputData.CustomPkg, util.GetConfig().DefaultPkgList.Packages...)
	specMap["format"] = insertData.BuildType
	specMap["architecture"] = insertData.Arch
	param := make(map[string]interface{})
	param["service"] = "omni"
	param["domain"] = "omni-build"
	param["task"] = "buildImage"
	param["engine"] = "kubernetes"
	param["userID"] = strconv.Itoa(insertData.UserId)
	param["spec"] = specMap
	paramBytes, _ := json.Marshal(param)
	result, err := util.HTTPPost(util.GetConfig().BuildServer.ApiUrl+"/v1/jobs", string(paramBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "HTTPPost Error", err))
		return
	}
	insertData.JobName = result["id"].(string)
	outputName := fmt.Sprintf(`openEuler-%s.iso`, result["id"])
	insertData.Status = result["state"].(string)
	insertData.StartTime, _ = time.Parse("2006-01-02T15:04:05Z", result["startTime"].(string))
	insertData.EndTime, _ = time.Parse("2006-01-02T15:04:05Z", result["endTime"].(string))
	insertData.DownloadUrl = fmt.Sprintf(util.GetConfig().BuildParam.DownloadIsoUrl, insertData.Release, time.Now().Format("2006-01-02"), outputName)
	insertData.Status = models.JOB_STATUS_START
	err = models.AddJobLog(&insertData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}

	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "构建OpenEuler"
	param["customRpms"] = imageInputData.CustomPkg
	delete(specMap, "packages")
	param["spec"] = specMap
	sd.Body = param
	sd.OperationTime = time.Now()
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, insertData))

}

// @Summary GetJobParam
// @Description get job build param
// @Tags  v2 job
// @Param	id		path 	string	true		"job id"
// @Accept json
// @Produce json
// @Router /v2/images/getJobParam/{id} [get]
func GetJobParam(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job id must be fill:", nil))
		return
	}
	result, err := models.GetJobLogByJobName(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "查询构建详情"
	sd.Body = id
	sd.OperationTime = time.Now()
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
}

// @Summary get single job detail
// @Description get single job detail
// @Tags  v2 job
// @Param	id		path 	string	true		"job id"
// @Accept json
// @Produce json
// @Router /v2/images/getOne/{id} [get]
func GetOne(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job id must be fill:", nil))
		return
	}
	param := url.Values{}
	param.Add("service", "omni")
	param.Add("domain", "omni-build")
	param.Add("task", "buildImage")
	param.Add("ID", id)
	result, err := util.HTTPGet(util.GetConfig().BuildServer.ApiUrl+"/v1/jobs", param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "查询构建日志"
	sd.Body = param
	sd.OperationTime = time.Now()
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))

}

// @Summary get single job logs
// @Description get single job logs
// @Tags  v2 job
// @Param	id		path 	string	true		"job id"
// @Param	stepID		query 	string	true		"stop id"
// @Param	uuid		query 	string	true		"uuid"
// @Accept json
// @Produce json
// @Router /v2/images/getLogsOf/{id} [get]
func GetJobLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job id must be fill:", nil))
		return
	}
	stepID, _ := strconv.Atoi(c.Query("stepID"))
	uuid := c.Query("uuid")
	param := url.Values{}
	param.Set("service", "omni")
	param.Set("domain", "omni-build")
	param.Set("task", "buildImage")
	param.Set("ID", id)
	param.Set("stepID", strconv.Itoa(stepID))
	if len(uuid) > 0 {
		param.Set("startTimeUUID", uuid)
	}

	param.Set("maxRecord", 999999999)
	var req *http.Request
	var err error
	req, err = http.NewRequest("GET", util.GetConfig().BuildServer.ApiUrl+"/v1/jobs/logs", nil)
	if param != nil {
		req.URL.RawQuery = param.Encode()
	}
	resp, err := http.DefaultClient.Do(req) //http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, stepID, err))
		return
	}
	defer resp.Body.Close()

	resultBytes, _ := ioutil.ReadAll(resp.Body)
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "查询构建日志详情"
	sd.Body = param
	sd.OperationTime = time.Now()
	result := string(resultBytes)
	if resp.StatusCode == 200 {
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
	} else {
		sd.State = "failed"
		sd.StateMessage = result
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, "error", result))
	}
}

// @Summary StopJobBuild
// @Description Stop Job Build
// @Tags  v2 job
// @Param	id		path 	string	true		"job id"
// @Accept json
// @Produce json
// @Router /v2/images/stopJob/{id} [delete]
func StopJobBuild(c *gin.Context) {
	id := c.Param("id")
	if len(id) < 10 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job id must be fill:", nil))
		return
	}

	param := url.Values{}
	param.Set("service", "omni")
	param.Set("domain", "omni-build")
	param.Set("task", "buildImage")
	param.Set("ID", id)
	param.Set("stepID", strconv.Itoa(stepID))
	param.Set("maxRecord", strconv.Itoa(maxRecord))
	var req *http.Request
	var err error
	req, err = http.NewRequest("GET", util.GetConfig().BuildServer.ApiUrl+"/v1/jobs/logs", nil)
	if param != nil {
		req.URL.RawQuery = param.Encode()
	}
	resp, err := http.DefaultClient.Do(req) //http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, stepID, err))
		return
	}
	defer resp.Body.Close()

	resultBytes, _ := ioutil.ReadAll(resp.Body)
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "查询构建日志详情"
	sd.Body = param
	sd.OperationTime = time.Now()
	result := string(resultBytes)
	if resp.StatusCode == 200 {
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
	} else {
		sd.State = "failed"
		sd.StateMessage = result
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, "error", result))
	}

}

// @Summary deleteRecord
// @Description delete a job build record
// @Tags  v2 job
// @Param	id		path 	string	true		"job id"
// @Accept json
// @Produce json
// @Router /v2/images/deleteJob/{id} [delete]
func DeleteJobLogs(c *gin.Context) {
	id := c.Param("id")
	if len(id) < 10 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, " job id must be fill:", nil))
		return
	}

	err := models.DeleteJobLogById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "error", err))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "删除构建历史"
	body := make(map[string]interface{})
	body["userID"] = sd.UserId
	body["jobID"] = id
	sd.Body = body
	sd.OperationTime = time.Now()
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", id))

}

// @Summary MySummary
// @Description get my summary
// @Tags  v2 job
// @Accept json
// @Produce json
// @Router /v2/images/getMySummary [get]
func GetMySummary(c *gin.Context) {
	userId, _ := strconv.Atoi((c.Keys["id"]).(string))
	result, err := models.CountSummaryStatus(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "error", err))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId = userId
	sd.EventType = "获取构建统计"
	body := make(map[string]interface{})
	body["userid"] = userId
	sd.Body = body
	sd.OperationTime = time.Now()
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))

}
