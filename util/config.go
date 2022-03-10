package util

import (
	"bufio"
	"encoding/json"
	"os"
)

type Config struct {
	AppName     string         `json:"app_name"`
	AppModel    string         `json:"app_model"`
	AppHost     string         `json:"app_host"`
	AppPort     int            `json:"app_port"`
	Database    DatabaseConfig `json:"database"`
	RedisConfig RedisConfig    `json:"redis_config"`
}

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

//Redis属性定义
type RedisConfig struct {
	Addr     string `json:"addr"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

func init() {
	if dir, err := os.Getwd(); err == nil {
		parseConfig(dir + "/conf/app.json")
	}
}

//获取
func GetConfig() *Config {
	return cfg
}

//全局变量 对外不可见
var cfg *Config = nil

func parseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		panic(err)

	}
	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
