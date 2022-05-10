package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/omnibuildplatform/omni-manager/models"
	"github.com/omnibuildplatform/omni-manager/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// @Summary ImportBaseImages
// @Description import  a image meta data
// @Tags  v3 version
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/baseImages/import [post]
func ImportBaseImages(c *gin.Context) {
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "import Base Images"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.BaseImages
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, sd.StateMessage))
		return
	}
	_, filename := path.Split(imageInputData.Url)
	extName := "iso"
	if strings.Contains(filename, ".") {
		splitList := strings.Split(filename, ".")
		extName = splitList[len(splitList)-1]
		if strings.Contains(extName, "?") {
			extName = strings.Split(extName, "?")[0]
		}
		if strings.Contains(extName, "#") {
			extName = strings.Split(extName, "#")[0]
		}
		if strings.Contains(extName, "&") {
			extName = strings.Split(extName, "&")[0]
		}
	} else {
		extName = "binary"
	}
	imageInputData.CreateTime = time.Now().In(util.CnTime)
	imageInputData.UserId, _ = strconv.Atoi(c.Keys["id"].(string))
	imageInputData.Status = models.ImageStatusStart
	imageInputData.ExtName = extName
	err = models.AddBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	//
	req, _ := http.NewRequest(http.MethodPost, util.GetConfig().BuildServer.OmniRepoAPI+"/data/loadfrom", nil)
	param := url.Values{}
	param.Add("url", imageInputData.Url)
	param.Add("userid", strconv.Itoa(imageInputData.UserId))
	param.Add("username", c.GetString("username"))
	param.Add("desc", imageInputData.Desc)
	param.Add("checksum", imageInputData.Checksum)
	param.Add("token", "316462d0c029ba707ad1")
	param.Add("externalID", strconv.Itoa(imageInputData.ID))
	param.Add("name", imageInputData.Name)
	req.URL.RawQuery = param.Encode()
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusServerError, "DefaultClient", err.Error()))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "ReadAll", err))
		return
	}
	if resp.StatusCode >= 400 {
		c.Data(http.StatusBadRequest, binding.MIMEJSON, respBody)
		return
	}

	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", imageInputData))

}

// @Summary RepoSavedCallBack
// @Description callback after repo loaded from source url
// @Tags  v3 version
// @Param	id		path 	int	true		"id for image item"
// @Param	status		query 	string	true		"status for image item"
// @Accept json
// @Produce json
// @Router /v3/baseImages/repoCallback/{id} [get]
func RepoSavedCallBack(c *gin.Context) {
	var imageInputData models.BaseImages
	imageInputData.Status = c.Query("status")
	imageInputData.ID, _ = strconv.Atoi(c.Param("id"))
	if imageInputData.ID <= 0 {
		util.Log.Errorf("RepoSavedCallBack Error, callback id is %v", c.Param("id"))
		return
	}

	err := models.UpdateBaseImagesStatus(&imageInputData)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

}

// @Summary UpdateBaseImages
// @Description update  a base  images data
// @Tags  v3 version
// @Param	id		path 	int	true		"id for  content"
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/baseImages/{id} [put]
func UpdateBaseImages(c *gin.Context) {

	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "修改BaseImages 数据"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.BaseImages
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "ShouldBindJSON", sd.StateMessage))
		return
	}
	imageInputData.ID, _ = strconv.Atoi(c.Param("id"))
	if imageInputData.ID <= 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "   id must be fill:", nil))
		return
	}

	err = models.UpdateBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "UpdateBaseImages", sd.StateMessage))
		return
	}
	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", imageInputData))

}

// @Summary DeletBaseImages
// @Description delete  a base  images data
// @Tags  v3 version
// @Param	id		path 	int	true		"id for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/baseImages/{id} [delete]
func DeletBaseImages(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "   id must be fill:", nil))
		return
	}

	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "删除BaseImages 数据"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	deleteNum, err := models.DeleteBaseImagesById(sd.UserId, id)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	sd.Body = id
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", sd.Body, deleteNum))
}

// @Summary BuildFromISO
// @Description build a image from iso
// @Tags  v3 version
// @Param	body		body 	models.BaseImagesKickStart	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v3/images/buildFromIso [post]
func BuildFromISO(c *gin.Context) {
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "从ISO构建OpenEuler"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.BaseImagesKickStart
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "ShouldBindJSON", sd.StateMessage))
		return
	}
	sd.Body = imageInputData

	baseImageID, _ := strconv.Atoi(imageInputData.BaseImageID)
	if baseImageID <= 0 {
		sd.State = "failed"
		sd.StateMessage = " BaseImageID must be number "
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "BaseImageID", imageInputData))
		return
	}
	var baseimage *models.BaseImages
	baseimage, err = models.GetBaseImagesByID(baseImageID)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "GetBaseImagesByID", sd.StateMessage))
		return
	}
	var insertData models.JobLog
	kickStartMap := make(map[string]interface{})

	imageInputData.KickStartContent = strings.ReplaceAll(imageInputData.KickStartContent, "\n", "")
	kickStartMap["content"] = imageInputData.KickStartContent
	kickStartMap["name"] = imageInputData.KickStartName

	imageMap := make(map[string]interface{})
	imageMap["name"] = baseimage.Name + "." + baseimage.ExtName
	imageMap["url"] = util.GetConfig().BuildServer.OmniRepoAPI + "/data/browse/" + baseimage.ExtName + "/" + baseimage.Checksum + "." + baseimage.ExtName
	// baseimage.Url
	imageMap["checksum"] = baseimage.Checksum
	imageMap["architecture"] = baseimage.Arch

	specMap := make(map[string]interface{})
	specMap["kickStart"] = kickStartMap
	specMap["image"] = imageMap
	param := make(map[string]interface{})
	param["service"] = "build"
	param["domain"] = "omni-build"
	param["task"] = models.BuildImageFromISO
	param["engine"] = "kubernetes"
	param["userID"] = strconv.Itoa(insertData.UserId)
	param["spec"] = specMap
	paramBytes, _ := json.Marshal(param)
	result, err := util.HTTPPost(util.GetConfig().BuildServer.ApiUrl+"/v1/jobs", string(paramBytes))
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "HTTPPost Error", err.Error()))
		return
	}

	paramBytes, _ = json.Marshal(result)
	fmt.Println("\n-------------------:", string(paramBytes))

	insertData.JobName = result["id"].(string)
	outputName := fmt.Sprintf(`%s.%s`, insertData.JobName, baseimage.ExtName)
	insertData.Status = result["state"].(string)
	insertData.StartTime, _ = time.Parse(time.RFC3339, result["startTime"].(string))
	insertData.EndTime, _ = time.Parse(time.RFC3339, result["endTime"].(string))
	insertData.UserName = c.Keys["nm"].(string)
	insertData.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	insertData.BuildType = "installer-iso"
	insertData.CreateTime = time.Now().In(util.CnTime)
	insertData.JobLabel = imageInputData.Label
	insertData.JobDesc = imageInputData.Desc
	insertData.JobType = models.BuildImageFromISO
	insertData.Arch = baseimage.Arch
	insertData.DownloadUrl = util.GetConfig().BuildServer.OmniRepoAPI + "/data/browse/" + baseimage.ExtName + "/" + outputName
	insertData.Status = models.JOB_STATUS_START
	err = models.AddJobLog(&insertData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "AddJobLog", sd.StateMessage))
		return
	}

	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", insertData))

}

// @Summary ListBaseImages
// @Description get my base image list order by id desc
// @Tags  v3 version
// @Param	offset		query 	int	false		"offset "
// @Param	limit		query 	int	false		"limit"
// @Accept json
// @Produce json
// @Router /v3/baseImages/list [get]
func ListBaseImages(c *gin.Context) {
	userId, _ := strconv.Atoi((c.Keys["id"]).(string))
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}

	total, imageList, err := models.GetMyBaseImages(userId, (offset), (limit))
	if err != nil {

		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", imageList, total))
}

// @Summary AddKickStart
// @Description add  a KickStart data
// @Tags  v3 version
// @Param file formData file true "kickstart file"
// @Param name formData string true "  name"
// @Param desc formData string true "  desc"
// @Accept json
// @Produce json
// @Router /v3/kickStart [post]
func AddKickStart(c *gin.Context) {
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "添加KickStart 数据"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.KickStart

	kickfile, err := c.FormFile("file")
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, sd.StateMessage))
		return
	}
	partFile, err := kickfile.Open()
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, sd.StateMessage))
		return
	}
	defer partFile.Close()

	fileContent, err := ioutil.ReadAll(partFile)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, sd.StateMessage))
		return
	}

	if models.IsUtf8(fileContent) == false {
		sd.State = "failed"
		sd.StateMessage = "kickstart file is not utf8 format"
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "err", sd.StateMessage))
		return
	}
	imageInputData.Content = string(fileContent)
	imageInputData.Name = c.Request.FormValue("name")
	imageInputData.Desc = c.Request.FormValue("desc")
	imageInputData.CreateTime = time.Now().In(util.CnTime)
	imageInputData.UserId, _ = strconv.Atoi(c.Keys["id"].(string))
	err = models.AddKickStart(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", &imageInputData))

}

// @Summary ListKickStart
// @Description get my kick start file list order by id desc
// @Tags  v3 version
// @Param	offset		query 	int	false		"offset "
// @Param	limit		query 	int	false		"limit"
// @Accept json
// @Produce json
// @Router /v3/kickStart/list [get]
func ListKickStart(c *gin.Context) {
	userId, _ := strconv.Atoi((c.Keys["id"]).(string))
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 10
	}

	total, dataList, err := models.GetMyKickStart(userId, (offset), (limit))
	if err != nil {

		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", dataList, total))
}

// @Summary UpdateKickStart
// @Description update  a kick start data
// @Tags  v3 version
// @Param	id		path 	int	true		"id for  content"
// @Param	body		body 	models.KickStart	true		"body for KickStart content"
// @Accept json
// @Produce json
// @Router /v3/kickStart/{id} [put]
func UpdateKickStart(c *gin.Context) {

	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "修改KickStart 数据"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.KickStart
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, sd.StateMessage))
		return
	}
	imageInputData.ID, _ = strconv.Atoi(c.Param("id"))
	if imageInputData.ID <= 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "   id must be fill:", nil))
		return
	}
	err = models.UpdateKickStart(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, sd.StateMessage))
		return
	}
	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", imageInputData))
}

// @Summary DeleteKickStart
// @Description delete  a KickStart data
// @Tags  v3 version
// @Param	id		path 	int	true		"id for KickStart content"
// @Accept json
// @Produce json
// @Router /v3/kickStart/{id} [delete]
func DeleteKickStart(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "   id must be fill:", nil))
		return
	}
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "删除KickStart 数据"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	deleteNum, err := models.DeleteKickStartById(sd.UserId, id)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	sd.Body = id
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", sd.Body, deleteNum))
}

// @Summary GetImagesAndKickStart
// @Description GetImagesAndKickStart
// @Tags  v3 version
// @Accept json
// @Produce json
// @Router /v3/getImagesAndKickStart [get]
func GetImagesAndKickStart(c *gin.Context) {

	userid, _ := strconv.Atoi((c.Keys["id"]).(string))
	result, err := models.GetImagesAndKickStart(userid)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))

}

// @Summary GetKickStartByID
// @Description GetKickStartByID
// @Tags  v3 version
// @Param	id		path 	int	true		"id for  content"
// @Accept json
// @Produce json
// @Router /v3/kickStart/{id} [get]
func GetKickStartByID(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))
	result, err := models.GetKickStartByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))
}
