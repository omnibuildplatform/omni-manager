package controllers

import (
	"net/http"
	"strconv"

	"github.com/omnibuildplatform/omni-manager/models"
	"github.com/omnibuildplatform/omni-manager/util"

	"github.com/gin-gonic/gin"
)

// @Summary AddBaseImages
// @Description add  a image meta data
// @Tags  v3 job
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/images/addBaseImages [post]
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
	err = models.AddBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

}

// @Summary UpdateBaseImages
// @Description update  a base  images data
// @Tags  v3 job
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/images/updateBaseImages [put]
func UpdateBaseImages(c *gin.Context) {

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
	err = models.AddBaseImages(&imageInputData)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

}

// @Summary DeletBaseImages
// @Description delete  a base  images data
// @Tags  v3 job
// @Param	body		body 	models.BaseImages	true		"body for BaseImages content"
// @Accept json
// @Produce json
// @Router /v3/images/deleteBaseImages [put]
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
	err := models.DeleteBaseImagesById(sd.UserId, id)
	if err != nil {
		sd.State = "failed"
		sd.StateMessage = err.Error()
		util.StatisticsLog(&sd)
		c.JSON(http.StatusBadRequest, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	sd.Body = id
	util.StatisticsLog(&sd)

}
