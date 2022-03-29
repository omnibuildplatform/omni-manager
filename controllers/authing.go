package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"omni-manager/models"
	"omni-manager/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type tokenItem struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// @Summary login success redirect url
// @Description login success redirect url
// @Tags  Authing
// @Accept json
// @Produce json
// @Router /v1/auth/loginok [get]
func AuthingLoginOk(c *gin.Context) {
	code, _ := c.GetQuery("code")

	// authorization_code
	resp, err := http.PostForm("https://openeuler-omni-manager.authing.cn/oidc/token",
		url.Values{
			"code":          {code},
			"client_id":     {"623d6bf75c72636ebb8c5e4b"},
			"client_secret": {"374a74fb1131916139d4789ef870820e"},
			"grant_type":    {"authorization_code"},
			"redirect_uri":  {"http://192.168.1.193:8080/api/v1/auth/loginok"}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "login err", err))
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "login err", err))
		return
	}
	var token tokenItem
	err = json.Unmarshal(body, &token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, body, err))
		return
	}

	jwtToken := new(jwt.Token)
	jwtToken.Valid = false
	jwtToken, err = jwt.Parse(token.AccessToken, func(jwtToken *jwt.Token) (interface{}, error) {
		return util.GetConfig().AuthingConfig.AppSecret, nil
	})

	// var jwtToken jwt.Token
	// if len(token.AccessToken) > 20 {
	// 	fmt.Println("------------jwtToken:", jwtToken)
	// 	fmt.Println("------------models.AuthingJWKSItem:", models.AuthingJWKSItem)

	// 	signStr, err := jwtToken.SignedString(models.AuthingJWKSItem)
	if err != nil {
		fmt.Println(util.GetConfig().AuthingConfig.AppSecret, "---------jwtToken Parse---:", err)
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "jwtToken Parse", err))
		return
	}
	// 	jwtToken.Method.Verify(token.AccessToken, signStr, map[string]string{
	// 		"issuer":   "https://openeuler-omni-manager.authing.cn/oidc",
	// 		"audience": "623d6bf75c72636ebb8c5e4b",
	// 	})

	if jwtToken.Valid {
		userinfo, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "jwt userinfo invalid", nil))
			return
		}
		c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, token, userinfo))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "token", token))
}

// @Summary AuthingGetToken
// @Description AuthingGetToken
// @Tags  Authing
// @Param	authingUserId		path 	string	true		"The key for staticblock"
// @Accept json
// @Produce json
// @Router /v1/auth/getDetail/{authingUserId} [get]
func AuthingGetUserDetail(c *gin.Context) {
	authingUserId := c.Param("authingUserId")
	userDetail, err := models.AuthingClient.Detail(authingUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "AuthingClient.Detail", err))
		return
	}
	if userDetail.Name == nil {
		defaultName := "noName"
		userDetail.Name = &defaultName
	}
	result := make(map[string]interface{})
	result["username"] = userDetail.Username
	result["nickname"] = userDetail.Nickname
	result["nm"] = userDetail.Name
	jwtString, err := models.GetJwtString(util.GetConfig().JwtConfig.Expire, userDetail.Id, *(userDetail.Name))
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "models.GetJwtString", err))
		return
	}
	result["token"] = jwtString
	result["photo"] = userDetail.Photo
	result["id"] = userDetail.Id
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", result))

}

// @Summary AuthingCreateUser
// @Description AuthingCreateUser
// @Tags  Authing
// @Param	body		body 	models.CreateUserInput	true		"body for user info"
// @Accept json
// @Produce json
// @Router /v1/auth/createUser [post]
func AuthingCreateUser(c *gin.Context) {
	token := c.GetHeader("token")
	if token == "" {
		c.JSON(http.StatusForbidden, util.ExportData(util.CodeStatusClientError, "forbidden", nil))
		return
	}
	var userInfo models.CreateUserInput
	err := c.BindJSON(&userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "error", err))
		return
	}

	newUserRequest := models.ParseAuthingUserInput(&userInfo)
	newuser, err := models.AuthingClient.CreateUser(*newUserRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.ExportData(util.CodeStatusServerError, "error", err))
		return
	}
	c.JSON(http.StatusOK, util.ExportData(util.CodeStatusNormal, "ok", newuser))

}
