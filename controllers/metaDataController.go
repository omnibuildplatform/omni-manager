package controllers

import (
	"encoding/json"
	"net/http"
	"omni-manager/models"
	"omni-manager/util"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary create
// @Description insert single data
// @Tags  meta Manager
// @Param	body		body 	models.ImageInputData	true		"body for Metadata content"
// @Accept json
// @Produce json
// @Router /images/insert [post]
func Insert(c *gin.Context) {
	var imageInputData models.ImageInputData
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	var insertData models.Metadata
	insertData.Architecture = imageInputData.Architecture
	insertData.EulerVersion = imageInputData.EulerVersion
	insertData.OutFormat = imageInputData.OutFormat
	var temp []byte
	temp, err = json.Marshal(imageInputData.BasePkg)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	insertData.BasePkg = string(temp)
	temp, err = json.Marshal(imageInputData.CustomPkg)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	insertData.CustomPkg = string(temp)
	insertData.CreateTime = time.Now()
	_, err = models.AddMetadata(&insertData)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", insertData))
}

// @Summary get
// @Description get single one
// @Tags  meta Manager
// @Param	id		path 	string	true		"The key for staticblock"
// @Accept json
// @Produce json
// @Router /images/get/{id} [get]
func Read(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if id <= 0 || err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, "id must be int type", err))
		return
	}

	v, err := models.GetMetadataById(id)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, id, v))
}

// @Summary query multi datas
// @Description use param to query multi datas
// @Tags  meta Manager
// @Param	project_name		query 	string	true		"project name"
// @Param	pkg_name		query 	string	true		"package name"
// @Accept json
// @Produce json
// @Router /images/query [get]
func Query(c *gin.Context) {
	//...... emplty . wait for query param
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, c.Query("project_name"), c.Query("pkg_name")))
}

// @Summary update
// @Description update single data
// @Tags  meta Manager
// @Param	body		body 	models.ImageInputData	true		"body for Metadata content"
// @Accept json
// @Produce json
// @Router /images/update [put]
func Update(c *gin.Context) {
	var imageInputData models.ImageInputData
	err := c.ShouldBindJSON(&imageInputData)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	var updateData models.Metadata
	updateData.Architecture = imageInputData.Architecture
	updateData.EulerVersion = imageInputData.EulerVersion
	updateData.OutFormat = imageInputData.OutFormat
	var temp []byte
	temp, err = json.Marshal(imageInputData.BasePkg)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	updateData.BasePkg = string(temp)
	temp, err = json.Marshal(imageInputData.CustomPkg)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, err, nil))
		return
	}
	updateData.CustomPkg = string(temp)
	//use origin item id
	updateData.Id = imageInputData.Id

	err = models.UpdateMetadataById(&updateData)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	util.Log.Warnf("The MetaData of Id (%d) had been update to: %s", updateData.Id, updateData.ToString())
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", nil))
}

// @Summary delete
// @Description update single data
// @Tags  meta Manager
// @Param	id		path 	string	true		"The key for staticblock"
// @Accept json
// @Produce json
// @Router /images/delete/:id [delete]
func Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if id <= 0 || err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusClientError, "id must be int type", err))
		return
	}
	err = models.DeleteMetadata(id)
	if err != nil {
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusServerError, err, nil))
		return
	}
	util.Log.Warnf("The  MetaData (Id:%d) had been delete ", id)
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", id))
}
