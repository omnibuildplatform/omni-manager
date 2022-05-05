package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
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
	fmt.Println("==========:", string(respBody))
	sd.Body = imageInputData
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", string(respBody)))

}

// @Summary AddBaseImages
// @Description add  a image meta data
// @Tags  v3 version
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/baseImages/add [post]
func AddBaseImages(c *gin.Context) {
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "添加BaseImages 数据"
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
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

}

// @Summary UpdateBaseImages
// @Description update  a base  images data
// @Tags  v3 version
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/baseImages/update [put]
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
// @Router /v3/baseImages/delete/{id} [delete]
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
// @Tags  v2 version
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
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	sd.Body = nil
	util.StatisticsLog(&sd)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

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
// @Param	body		body 	models.KickStart	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/kickStart/add [post]
func AddKickStart(c *gin.Context) {
	sd := util.StatisticsData{}
	sd.UserId, _ = strconv.Atoi((c.Keys["id"]).(string))
	sd.EventType = "添加KickStart 数据"
	if c.Keys["p"] != nil {
		sd.UserProvider = (c.Keys["p"]).(string)
	}
	var imageInputData models.KickStart
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
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

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
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
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
// @Param	body		body 	models.KickStart	true		"body for KickStart content"
// @Accept json
// @Produce json
// @Router /v3/kickStart/update [put]
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
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
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

	err = models.UpdateKickStart(&imageInputData)
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

// @Summary DeleteKickStart
// @Description delete  a KickStart data
// @Tags  v3 version
// @Param	id		path 	int	true		"id for KickStart content"
// @Accept json
// @Produce json
// @Router /v3/kickStart/delete/{id} [delete]
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

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// @Summary testItem
// @Description add  a image meta data
// @Tags  v3 version
// @Param	id		query 	int	true		"int"
// @Accept json
// @Produce json
// @Router /v3/baseImages/testItem [get]
func TestItem(c *gin.Context) {

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, i, nil))

	}
}
