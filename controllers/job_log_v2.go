package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/omnibuildplatform/omni-manager/models"
	"github.com/omnibuildplatform/omni-manager/util"

	"github.com/gin-gonic/gin"
)

// @Summary Create Job
// @Description start a image build job
// @Tags  v2 version
// @Param	body		body 	models.BuildParam	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v2/images/createJob [post]
func CreateJob(c *gin.Context) {
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "构建OpenEuler"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.BuildParam
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
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
	insertData.JobType = models.BuildImageFromRelease
	insertData.JobDesc = imageInputData.Desc
	if insertData.JobLabel == "" {
		insertData.JobLabel = insertData.UserName + "_" + insertData.Arch + "_" + insertData.Release
	}
	if insertData.JobDesc == "" {
		insertData.JobDesc = "this image was built by Omni Build Platform"
	}
	insertData.CreateTime = time.Now().In(util.CnTime)
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	if len(insertData.Release) == 0 {
		sd.State = "failed"
		sd.StateMessage = "release is empty"
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "Release not allowed empty ", nil))
		return
	}

	if insertData.UserId <= 0 {
		sd.State = "failed"
		sd.StateMessage = "user is not right"
		util.StatisticsLog(&sd)
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
		sd.State = "failed"
		sd.StateMessage = "arch is not right"
		util.StatisticsLog(&sd)
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
		sd.State = "failed"
		sd.StateMessage = "build type is not right"
		util.StatisticsLog(&sd)
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
	param["task"] = models.BuildImageFromRelease
	param["engine"] = "kubernetes"
	param["userID"] = strconv.Itoa(insertData.UserId)
	param["spec"] = specMap
	paramBytes, _ := json.Marshal(param)
	delete(specMap, "packages")
	result, err := util.HTTPPost(util.GetConfig().BuildServer.ApiUrl+"/v1/jobs", string(paramBytes))
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "HTTPPost Error", err.Error()))
		return
	}
	// util.Log.Debug(util.GetConfig().BuildServer.ApiUrl, "v2 CreateJob:------------:", result)
	insertData.JobName = result["id"].(string)
	outputName := fmt.Sprintf(`openEuler-%s.iso`, result["id"])
	insertData.Status = result["state"].(string)
	insertData.StartTime, _ = time.Parse(time.RFC3339, result["startTime"].(string))
	insertData.EndTime, _ = time.Parse(time.RFC3339, result["endTime"].(string))
	insertData.DownloadUrl = util.GetConfig().BuildServer.OmniRepoAPI + "/data/browse/" + insertData.Release + "/" + time.Now().In(util.CnTime).Format("2006-01-02") + "/" + outputName
	insertData.Status = models.JOB_STATUS_START
	err = models.AddJobLog(&insertData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}
	param["customRpms"] = imageInputData.CustomPkg
	sd.Body = param
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, insertData))

}

// @Summary GetJobParam
// @Description get job build param
// @Tags  v2 version
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
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
}

// @Summary get single job detail
// @Description get single job detail
// @Tags  v2 version
// @Param	id		path 	string	true		"job id"
// @Param	jobtype		query 	string	true		"job type"
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
	jobtype := c.Query("jobtype")
	if len(jobtype) == 0 {
		jobtype = models.BuildImageFromRelease
	}
	param.Add("task", jobtype)
	param.Add("ID", id)
	result, err := util.HTTPGet(util.GetConfig().BuildServer.ApiUrl+"/v1/jobs", param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, nil, err))
		return
	}

	if result["state"] == models.JOB_BUILD_STATUS_SUCCEED {
		if result["task"] == models.BuildImageFromISO {
			downloadURL := util.GetConfig().BuildServer.OmniRepoAPI + "/data/query?externalID=" + result["id"].(string)

			c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result, downloadURL))

		} else {
			c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
		}
	} else {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
	}

}

// @Summary get single job logs
// @Description get single job logs
// @Tags  v2 version
// @Param	id		path 	string	true		"job id"
// @Param	stepID		query 	string	true		"step id"
// @Param	uuid		query 	string	false		"uuid"
// @Param	jobtype		query 	string	true		"job type"
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
	jobtype := c.Query("jobtype")
	if len(jobtype) == 0 {
		jobtype = models.BuildImageFromRelease
	}
	param.Set("task", jobtype)
	param.Set("ID", id)
	param.Set("stepID", strconv.Itoa(stepID))
	if len(uuid) > 0 {
		param.Set("startTimeUUID", uuid)
	}
	param.Set("maxRecord", "999999999")
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
	Logcompleted := resp.Header["Logcompleted"]
	Logtimeuuid := resp.Header["Logtimeuuid"]
	resultBytes, _ := ioutil.ReadAll(resp.Body)
	result := make(map[string]string)
	if len(Logtimeuuid) > 0 {
		result["uuid"] = Logtimeuuid[0]
	}
	if len(Logcompleted) > 0 {
		result["stopOK"] = Logcompleted[0]
	}
	result["log"] = string(resultBytes)
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "查询构建日志"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	sd.Body = param
	if resp.StatusCode == 200 {
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
	} else {
		sd.State = "failed"
		sd.StateMessage = result["log"]
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, "error", result))
	}
}

// @Summary StopJobBuild
// @Description Stop Job Build
// @Tags  v2 version
// @Param	id		path 	string	true		"job id"
// @Param	jobtype		query 	string	true		"job type"
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
	jobtype := c.Query("jobtype")
	if len(jobtype) == 0 {
		jobtype = models.BuildImageFromRelease
	}
	param.Set("task", jobtype)
	param.Set("ID", id)
	var req *http.Request
	var err error
	req, err = http.NewRequest("DELETE", util.GetConfig().BuildServer.ApiUrl+"/v1/jobs", nil)
	if param != nil {
		req.URL.RawQuery = param.Encode()
	}
	resp, err := http.DefaultClient.Do(req) //http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "error", err))
		return
	}
	defer resp.Body.Close()

	resultBytes, _ := ioutil.ReadAll(resp.Body)
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "stop构建过程"
	sd.Body = param
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	result := string(resultBytes)
	if resp.StatusCode == 200 {
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
	} else {
		sd.State = "failed"
		sd.StateMessage = result
		util.StatisticsLog(&sd)
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusServerError, "error", result))
	}

}

// @Summary deleteRecord
// @Description delete multipule job build records
// @Tags  v2 version
// @Param	body		body 	[]string	true		"job id list"
// @Accept json
// @Produce json
// @Router /v2/images/deleteJob [post]
func DeleteJobLogs(c *gin.Context) {

	var nameList []string
	err := c.ShouldBindJSON(&nameList)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "error", err))
		return
	}
	if len(nameList) == 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "error", "id list must fill"))
		return
	}
	var names string
	for i, name := range nameList {
		if i == 0 {
			names = "'" + name + "'"
		} else {
			names = names + ",'" + name + "'"
		}
	}
	err = models.DeleteMultiJobLogs(names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "error", err))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "删除构建历史"
	body := make(map[string]interface{})
	body["userID"] = sd.UserId
	body["jobID"] = nameList
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	sd.Body = body
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", nameList))

}

// @Summary MySummary
// @Description get my summary
// @Tags  v2 version
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
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))

}
