package util

import (
	"bufio"
	"encoding/json"
	"os"
)

//app config
type Config struct {
	AppName     string         `json:"app_name"`
	AppModel    string         `json:"app_model"`
	AppHost     string         `json:"app_host"`
	AppPort     int            `json:"app_port"`
	Database    DatabaseConfig `json:"database"`
	RedisConfig RedisConfig    `json:"redis_config"`
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

func init() {
	//app.json must be set right folder
	if dir, err := os.Getwd(); err == nil {
		parseConfig(dir + "/conf/app.json")
	}
}

//external
func GetConfig() *Config {
	return cfg
}

//internal
var cfg *Config = nil

func parseConfig(path string) (*Config, error) {
	file, err := os.Open(path)

	if err != nil {
		Log.Errorf("read config file failed, please check path .  app exit now .")
		os.Exit(1)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
