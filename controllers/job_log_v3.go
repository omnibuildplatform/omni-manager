package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	imageInputData.CreateTime = time.Now().In(util.CnTime)
	imageInputData.UserId, _ = strconv.Atoi(c.Keys["id"].(string))
	req, _ := http.NewRequest(http.MethodPost, util.GetConfig().BuildServer.ImagesRepoAPI+"/data/loadfrom", nil)
	param := url.Values{}
	param.Add("url", imageInputData.Url)
	param.Add("userid", strconv.Itoa(imageInputData.UserId))
	param.Add("username", c.GetString("username"))
	param.Add("desc", c.GetString("desc"))
	param.Add("checksum", imageInputData.Checksum)
	param.Add("token", "316462d0c029ba707ad1")
	req.URL.RawQuery = param.Encode()
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	if resp.StatusCode != 200 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusServerError, "ReadAll err", respBody))
		return
	}
	imageInputData.Status = "downloading"
	err = models.AddBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", string(respBody)))

}

// // @Summary AddBaseImages
// // @Description add  a image meta data
// // @Tags  v3 version
// // @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// // @Accept json
// // @Produce json
// // @Router /v3/baseImages/add [post]
// func AddBaseImages(c *gin.Context) {
// 	sd := util.StatisticsData{}
// 	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
// 	sd.EventType = "添加BaseImages 数据"
// 	if c.Keys["p"] != nil {
// 		sd.UserProvider = (c.Keys["p"]).(string)
// 	}
// 	var imageInputData models.BaseImages
// 	err := c.ShouldBindJSON(&imageInputData)
// 	if err != nil {
// 		sd.State = "failed"
// 		sd.StateMessage = err.Error()
// 		util.StatisticsLog(&sd)
// 		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
// 		return
// 	}
// 	imageInputData.CreateTime = time.Now().In(util.CnTime)
// 	imageInputData.UserId, _ = strconv.Atoi(c.Keys["id"].(string))
// 	err = models.AddBaseImages(&imageInputData)
// 	if err != nil {
// 		sd.State = "failed"
// 		sd.StateMessage = err.Error()
// 		util.StatisticsLog(&sd)
// 		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
// 		return
// 	}
// 	sd.Body = imageInputData
// 	util.StatisticsLog(&sd)
// 	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

// }

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
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	imageInputData.ID, _ = strconv.Atoi(c.Param("id"))
	if imageInputData.ID <= 0 {
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, "   id must be fill:", nil))
		return
	}
	userid, _ := strconv.Atoi(c.Keys["id"].(string))
	if userid != imageInputData.UserId {
		sd.State = "failed"
		sd.StateMessage = fmt.Sprintf("this one:[%d] is not auther[%d]", userid, imageInputData.UserId)
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	err = models.UpdateBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
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
	insertData.UserName = c.Keys["nm"].(string)
	insertData.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	insertData.BuildType = models.BuildImageFromISO
	insertData.CreateTime = time.Now().In(util.CnTime)
	insertData.JobLabel = imageInputData.Name
	insertData.JobDesc = imageInputData.Desc
	insertData.Arch = baseimage.Arch
	insertData.DownloadUrl = "" // fmt.Sprintf(util.GetConfig().BuildParam.DownloadIsoUrl, baseimage.Name, time.Now().In(util.CnTime).Format("2006-01-02"), outputName)
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
