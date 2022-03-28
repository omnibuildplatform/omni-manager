package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"omni-manager/util"

	"github.com/Authing/authing-go-sdk/lib/authentication"
	"github.com/Authing/authing-go-sdk/lib/management"
	"github.com/Authing/authing-go-sdk/lib/model"
	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Username          *string  `json:"username,omitempty"`
	Email             *string  `json:"email,omitempty"`
	EmailVerified     *bool    `json:"emailVerified,omitempty"`
	Phone             *string  `json:"phone,omitempty"`
	PhoneVerified     *bool    `json:"phoneVerified,omitempty"`
	Unionid           *string  `json:"unionid,omitempty"`
	Openid            *string  `json:"openid,omitempty"`
	Nickname          *string  `json:"nickname,omitempty"`
	Photo             *string  `json:"photo,omitempty"`
	Password          *string  `json:"password,omitempty"`
	RegisterSource    []string `json:"registerSource,omitempty"`
	Browser           *string  `json:"browser,omitempty"`
	Oauth             *string  `json:"oauth,omitempty"`
	LoginsCount       *int64   `json:"loginsCount,omitempty"`
	LastLogin         *string  `json:"lastLogin,omitempty"`
	Company           *string  `json:"company,omitempty"`
	LastIP            *string  `json:"lastIP,omitempty"`
	SignedUp          *string  `json:"signedUp,omitempty"`
	Blocked           *bool    `json:"blocked,omitempty"`
	IsDeleted         *bool    `json:"isDeleted,omitempty"`
	Device            *string  `json:"device,omitempty"`
	Name              *string  `json:"name,omitempty"`
	GivenName         *string  `json:"givenName,omitempty"`
	FamilyName        *string  `json:"familyName,omitempty"`
	MiddleName        *string  `json:"middleName,omitempty"`
	Profile           *string  `json:"profile,omitempty"`
	PreferredUsername *string  `json:"preferredUsername,omitempty"`
	Website           *string  `json:"website,omitempty"`
	Gender            *string  `json:"gender,omitempty"`
	Birthdate         *string  `json:"birthdate,omitempty"`
	Zoneinfo          *string  `json:"zoneinfo,omitempty"`
	Locale            *string  `json:"locale,omitempty"`
	Address           *string  `json:"address,omitempty"`
	Formatted         *string  `json:"formatted,omitempty"`
	StreetAddress     *string  `json:"streetAddress,omitempty"`
	Locality          *string  `json:"locality,omitempty"`
	Region            *string  `json:"region,omitempty"`
	PostalCode        *string  `json:"postalCode,omitempty"`
	Country           *string  `json:"country,omitempty"`
	ExternalId        *string  `json:"externalId,omitempty"`
}
type AuthingKey struct {
	E   string `json:"e"`
	N   string `json:"n"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"Kid"`
}

type AuthingJWKS struct {
	Keys []AuthingKey
}

var AuthingClient *management.Client
var AppClient *model.Application
var UserClient *authentication.Client
var AuthingJWKSItem AuthingJWKS

func InitAuthing(userpoolid, secret string) {
	if userpoolid == "" {
		userpoolid = util.GetConfig().AuthingConfig.UserPoolID
		secret = util.GetConfig().AuthingConfig.Secret
	}
	AuthingClient = management.NewClient(userpoolid, secret)
	AppClient, _ = AuthingClient.FindApplicationById("623d6bf75c72636ebb8c5e4b")
	UserClient = authentication.NewClient(util.GetConfig().AuthingConfig.AppID, util.GetConfig().AuthingConfig.Secret)
	return
	resp, err := http.Get("https://openeuler-omni-manager.authing.cn/oidc/.well-known/jwks.json")
	if err != nil {
		fmt.Println("----------Get jwks error:", err)
		return
	}
	defer resp.Body.Close()
	//-------------------------------

	//--------------------------jwt
	jkwsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("----------Get jwks ReadAll error:", err)
		return
	}
	err = json.Unmarshal(jkwsBytes, &AuthingJWKSItem)
	if err != nil {
		fmt.Println("----------Get jwks Unmarshal error:", err)
		return
	}
	fmt.Println("--------", AuthingJWKSItem)

	fmt.Println("==========c.host:", AuthingClient.Host)
	userpoolDetai, _ := AuthingClient.UserPoolDetail()
	temBytes, err := json.Marshal(userpoolDetai)
	if err != nil {
		fmt.Println("----------UserPoolDetail---err---", err)
		return
	}
	fmt.Println("----------UserPoolDetail:", string(temBytes))
	comm := new(model.CommonPageRequest)
	comm.Limit = 100
	applistData, _ := AuthingClient.ListApplication(comm)
	for index, v := range applistData.List {
		fmt.Println(index, "-----ListApplication----:", v)
	}
	myqpp, _ := AuthingClient.CreateApplication("luonan App2", "luonancomapp2", "www.luonan2.net", nil)
	temBytes, err = json.Marshal(myqpp)
	fmt.Println("-----myqpp----:", string(temBytes))

	var testreq model.QueryListRequest
	testreq.Limit = 10
	pageUser, err := AuthingClient.GetUserList(testreq)
	if err != nil {
		fmt.Println("----------GetUserList---err---", err)
		return
	}
	fmt.Println("=========GetUserList:=====", pageUser)
	for index, user := range pageUser.List {
		fmt.Println(index, "-----user----:", user)
	}
	fmt.Println("--------------test cureate user")
	var testuser model.CreateUserRequest
	username := "luonancom2"
	testuser.UserInfo.Username = &username
	email := "582435826@qq.com"
	testuser.UserInfo.Email = &email
	newuser, err := AuthingClient.CreateUser(testuser)
	if err != nil {
		fmt.Println("----------CreateUser---err---", err)
		return
	}

	temBytes, err = json.Marshal(newuser)
	if err != nil {
		fmt.Println("----------UserPoolDetail---err---", err)
		return
	}

	fmt.Println("-----------new user:", string(temBytes))
	_, _ = resp, err
	AuthingClient.UserPoolDetail()
}
func ParseAuthingUserInput(userinput *CreateUserInput) *model.CreateUserRequest {
	fmt.Println(*(userinput.Address), "------ParseAuthingUserInput----------", userinput)

	var testuser model.CreateUserRequest
	testuser.UserInfo.Username = userinput.Username

	testuser.UserInfo.Email = userinput.Email

	return &testuser
}

func GetUserInfoByToekn(token string) error {
	resp, err := http.Get("https://openeuler-omni-manager.authing.cn/oidc/me?access_token=" + token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respDataBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("--------GetUserInfoByToekn==============", string(respDataBytes))
	return nil

}

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		loginStatus, err := UserClient.CheckLoginStatus(token)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusUnauthorized, util.ExportData(util.CodeStatusClientError, "forbidden", nil))
			return
		}
		if loginStatus.Status {
			//pass
			c.Next()
		} else {
			// no pass
			c.Abort()
			c.JSON(http.StatusUnauthorized, util.ExportData(util.CodeStatusClientError, "forbidden", nil))

			return
		}
	}
}
