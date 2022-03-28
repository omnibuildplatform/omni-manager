package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

//app config
type Config struct {
	AppName        string         `json:"app_name"`
	AppModel       string         `json:"app_model"`
	AppHost        string         `json:"app_host"`
	AppPort        int            `json:"app_port"`
	Database       DatabaseConfig `json:"database"`
	RedisConfig    RedisConfig    `json:"redis_config"`
	BuildParam     BuildParam     `json:"buildParam"`
	DefaultPkgList PkgList
	CustomPkgList  PkgList
	WSConfig       WSConfig      `json:"ws_config"`
	K8sConfig      K8sConfig     `json:"k8s"`
	AuthingConfig  AuthingConfig `json:"authing"`
}

type K8sConfig struct {
	Namespace string `json:"namespace"`
	Image     string `json:"image"`
	FfileType string `json:"ffileType"`
}

//sql config
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DbName   string `json:"db_name"`
	Chartset string `json:"charset"`
	ShowSql  bool   `json:"show_sql"`
}

//Redis config
type RedisConfig struct {
	Addr     string `json:"addr"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

//
type BuildParam struct {
	Arch             []string `json:"arch"`
	Release          []string `json:"release"`
	BuildType        []string `json:"buildType"`
	OpeneulerMinimal string   `json:"openeulerMinimal"`
	CustomRpmAPI     string   `json:"customRpmAPI"`
	PackageName      string   `json:"packageName"`
	DownloadIsoUrl   string   `json:"downloadIsoUrl"`
}
type PkgList struct {
	Packages []string `json:"packages"`
}

//websocket config
type WSConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	CheckOrigin bool   `json:"check_origin"`
}

//Authing Config
type AuthingConfig struct {
	UserPoolID string `json:"userPoolID"`
	Secret     string `json:"secret"`
	AppID      string `json:"appID"`
	AppSecret  string `json:"appSecret"`
}

func InitConfig() {
	//app.json must be set right folder
	if dir, err := os.Getwd(); err == nil {
		dir = dir + "/conf/app.json"
		err = parseConfig(dir)
		if err != nil {
			Log.Errorf("load app.json   failed, app must exit .please check app.json path:%s,and error:%s", dir, err)
			os.Exit(1)
			return
		}
	}
	// load openeuler_minimal.json file from github resp, and reload and update it'data every night at 3:00 / beijing
	minimalPath := fmt.Sprintf(GetConfig().BuildParam.OpeneulerMinimal, GetConfig().BuildParam.PackageName)
	respo, err := http.Get(minimalPath)
	if err != nil {
		Log.Errorf("load openEuler-minimal.json file failed, app must exit .please check url path:%s. and error:%s", minimalPath, err)
		os.Exit(1)
		return
	}
	defer respo.Body.Close()

	defaultPkg, err := ioutil.ReadAll(respo.Body)
	if err != nil {
		Log.Errorf("read data from %s file failed.err:%s", minimalPath, err)
		os.Exit(1)
		return
	}
	err = json.Unmarshal(defaultPkg, &(GetConfig().DefaultPkgList))
	if err != nil {
		Log.Errorf("config default package list is not json format :%s", err)
		os.Exit(1)
		return
	}
}

//external
func GetConfig() *Config {
	return cfg
}

//internal
var cfg *Config = nil

func parseConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		Log.Errorf("read config file failed, please check path .  app exit now .")
		os.Exit(1)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&cfg); err != nil {
		return err
	}
	return nil
}
