package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/omnibuildplatform/omni-manager/models"
	"github.com/omnibuildplatform/omni-manager/util"

	"github.com/gin-gonic/gin"
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
	if imageInputData.Algorithm == "" || imageInputData.Checksum == "" || imageInputData.Url == "" {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "err", "input param is not allowed be empty"))
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
	imageInputData.Checksum = strings.ToLower(imageInputData.Checksum)
	if imageInputData.Algorithm == "" {
		imageInputData.Algorithm = "sha256"
	}
	err = models.AddBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	param := make(map[string]interface{})
	param["sourceUrl"] = imageInputData.Url
	param["userID"] = imageInputData.UserId
	param["fileName"] = filename
	param["desc"] = imageInputData.Desc
	param["checksum"] = imageInputData.Checksum
	param["algorithm"] = imageInputData.Algorithm
	param["externalID"] = strconv.Itoa(imageInputData.ID)
	param["name"] = imageInputData.Name
	param["publish"] = false
	param["externalComponent"] = util.GetConfig().AppName

	bodyBytes, _ := json.Marshal(param)
	requestURL := util.GetConfig().BuildServer.OmniRepoAPIInternal + "/images/load"
	req, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		imageInputData.Status = models.ImageStatusFailed
		sd.StateMessage = "使用repo下载失败原因是:" + err.Error()
		models.UpdateBaseImagesStatus(&imageInputData)

		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		imageInputData.Status = models.ImageStatusFailed
		sd.StateMessage = "使用repo下载失败原因是:" + err.Error()
		models.UpdateBaseImagesStatus(&imageInputData)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusServerError, "DefaultClient", err.Error()))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "ReadAll", err))
		return
	}

	title := "ok"
	if resp.StatusCode >= 400 {
		fmt.Println("req:", req)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "error,url:"+requestURL, string(respBody)))
		return
	} else if resp.StatusCode == http.StatusAlreadyReported {
		imageInputData.Status = models.ImageStatusDone
		title = models.ImageStatusDone
		sd.StateMessage = "使用已经存在的image"
		models.UpdateBaseImagesStatus(&imageInputData)

	}
	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, title, imageInputData))

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
	//------------------delete image file from omni-repository
	baseImages, err := models.GetBaseImagesByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "GetBaseImagesByID error", err))
		return
	}

	req, err := http.NewRequest(http.MethodDelete, util.GetConfig().BuildServer.OmniRepoAPIInternal+"/images", nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	q := req.URL.Query()
	q.Add("userID", strconv.Itoa(baseImages.UserId))
	q.Add("checksum", baseImages.Checksum)
	req.URL.RawQuery = q.Encode()
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
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
		// go on to delete it if 404
		if resp.StatusCode != 404 {
			c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "error,url:"+req.URL.RawQuery, string(respBody)))
			return
		}
	}

	//-------------------------------delte record from database
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

	//--------------------------------
	//query really baseImage url before call build API.
	var req *http.Request
	req, _ = http.NewRequest("GET", util.GetConfig().BuildServer.OmniRepoAPI+"/images/query?externalID="+strconv.Itoa(baseimage.ID), nil)

	resp, err := http.DefaultClient.Do(req) //http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "GetRepositoryDownload", err))
		return
	}
	defer resp.Body.Close()
	resultBytes, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "qeury respository", string(resultBytes)))
		return
	}
	var imageResp models.ImageResponse
	err = json.Unmarshal(resultBytes, &imageResp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "json.Unmarshal error ", err))
		return
	}
	//-------------------------------
	baseimage.Checksum = strings.ToLower(baseimage.Checksum)
	var insertData models.JobLog
	kickStartMap := make(map[string]interface{})

	// imageInputData.KickStartContent = strings.ReplaceAll(imageInputData.KickStartContent, "\n", "  ")
	kickStartMap["content"] = strings.ReplaceAll(imageInputData.KickStartContent, "\n", "  ") // imageInputData.KickStartContent
	kickStartMap["name"] = imageInputData.KickStartName

	imageMap := make(map[string]interface{})
	imageMap["name"] = imageResp.Name
	imageMap["url"] = util.GetConfig().BuildServer.OmniRepoAPIInternal + imageResp.ImagePath //util.GetConfig().BuildServer.OmniRepoAPIInternal + "/images/query?externalID=" + strconv.Itoa(baseimage.ID)
	imageMap["checksum"] = imageResp.Checksum
	imageMap["architecture"] = baseimage.Arch

	specMap := make(map[string]interface{})
	specMap["kickStart"] = kickStartMap
	specMap["image"] = imageMap
	param := make(map[string]interface{})
	param["service"] = "omni"
	param["domain"] = "omni-build"
	param["task"] = models.BuildImageFromISO
	param["engine"] = "kubernetes"
	param["userID"] = (c.Keys["id"]).(string)
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

	insertData.JobName = result["id"].(string)
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
	insertData.KickStartID = imageInputData.KickStartID
	insertData.BaseImageID = imageInputData.BaseImageID
	insertData.KickStartContent = imageInputData.KickStartContent
	if insertData.JobLabel == "" {
		insertData.JobLabel = insertData.UserName + "_" + insertData.Arch + "_" + insertData.Release
	}
	if insertData.JobDesc == "" {
		insertData.JobDesc = "this image was built from custom baseImages "
	}

	insertData.Status = models.JOB_STATUS_CREATED
	err = models.AddJobLog(&insertData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "AddJobLog", sd.StateMessage))
		return
	}

	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", insertData, param))

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

	var unVerfiyImageList []*models.BaseImages
	for _, image := range imageList {
		if image.Status == string(models.ImageCreated) {
			// key := fmt.Sprintf("imageStatus:%s:%s", , externalItems[1])
			// util.GetFloat(key)

			unVerfiyImageList = append(unVerfiyImageList, image)
		}
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

// @Summary GetRepositoryDownlad
// @Description GetRepositoryDownlad
// @Tags  v3 version
// @Param	id		path 	int	true		"id for  content"
// @Accept json
// @Produce json
// @Router /v3/getRepositoryDownlad/{id} [get]
func GetRepositoryDownlad(c *gin.Context) {

	jobname := c.Param("id")
	if len(jobname) == 0 {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusClientError, "externalid must fill in", jobname))
		return
	}
	var req *http.Request
	req, _ = http.NewRequest("GET", util.GetConfig().BuildServer.OmniRepoAPI+"/images/query?externalID="+jobname, nil)

	resp, err := http.DefaultClient.Do(req) //http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "GetRepositoryDownlad", err))
		return
	}
	defer resp.Body.Close()
	resultBytes, _ := ioutil.ReadAll(resp.Body)

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, util.GetConfig().BuildServer.OmniRepoAPI, string(resultBytes)))
}
