package controllers

import (
	"net/http"
	"omni-manager/util"

	"github.com/gin-gonic/gin"
)

// @Summary AddBaseImages
// @Description add  a image image data
// @Tags  v3 job
// @Param	body		body 	models.BuildParam	true		"body for ImageMeta content"
// @Accept json
// @Produce json
// @Router /v3/images/addBaseImages [post]
func AddBaseImages(c *gin.Context) {

	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, 0, nil))

}
