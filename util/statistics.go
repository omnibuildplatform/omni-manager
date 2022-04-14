package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type StatisticsData struct {
	UserId        int
	UserName      string
	UserEmail     string
	OperationTime string
	EventType     string
	Value         string
	State         string
	StateMessage  string
	Body          string
}

func init() {

}

var statisticsLogFile *os.File

func dataFormatConver(sd StatisticsData) []byte {
	mapData := make(map[string]interface{})
	mapData["operationTime"] = fmt.Sprintf("%v", sd.OperationTime)
	mapData["userId"] = fmt.Sprintf("%v", sd.UserId)
	mapData["userName"] = fmt.Sprintf("%v", sd.UserName)
	mapData["eventType"] = fmt.Sprintf("%v", sd.EventType)
	mapData["value"] = fmt.Sprintf("%v", sd.Value)
	mapData["body"] = fmt.Sprintf("%v", sd.Body)
	mapData["appId"] = GetConfig().AuthingConfig.AppID
	data, err := json.Marshal(mapData)
	if err != nil {
		Log.Error("err: ", err)
	}
	return []byte(data)
}
func writeStatistLog(filePath string, byteData []byte) error {
	var err error
	statisticsLogFile, err = os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0600)
	defer statisticsLogFile.Close()
	if err != nil {
		Log.Error("fail to open the file, err: ", err, ",filePath: ", filePath)
		return err
	}
	_, err = statisticsLogFile.Write(byteData)
	_, err = statisticsLogFile.Write([]byte("\n"))
	return nil
}

func createStatistLog(logFile string) (string, error) {
	configPath := GetConfig().Statistic.Dir
	CreateDir(configPath)
	if len(logFile) == 0 {
		logFile = GetCurDate() + "_" + GetConfig().Statistic.LogFile
	}
	filePath := filepath.Join(configPath, logFile)
	if !FileExists(filePath) {
		f, err := os.Create(filePath)
		if err != nil {
			Log.Error("Failed to create file, err: ", err, ",filePath: ", filePath)
			return "", err
		}
		defer f.Close()
	}
	return filePath, nil
}

func convertStrToInt(num string) int64 {
	intNum, _ := strconv.ParseInt(num, 10, 64)
	return intNum
}

func renameStatistLog(filePath string) error {
	dir := GetConfig().Statistic.Dir
	fileSuffix := GetConfig().Statistic.LogFileSuffix
	files, _ := ioutil.ReadDir(dir)
	if len(files) > 0 {
		fileName := ""
		nameList := make([]string, 0)
		for _, f := range files {
			nameList = append(nameList, f.Name())
		}
		sort.Strings(nameList)
		lastFile := nameList[len(nameList)-1]
		splitFile := strings.Split(lastFile, ".log")
		if len(splitFile) < 2 {
			fileName = lastFile + fileSuffix
		} else {
			intNum := convertStrToInt(splitFile[1]) + 1
			format := "%0" + strconv.Itoa(len(fileSuffix)) + "d"
			fileName = lastFile + fmt.Sprintf(format, intNum)
		}
		err := os.Rename(filePath, fileName)
		if err != nil {
			Log.Error("file renaming failed, ", filePath, "====>", fileName)
			return err
		}
		createStatistLog(filePath)
	}
	return nil
}

func splitStatistLog(filePath string) error {
	f, err := os.Stat(filePath)
	if err != nil {
		Log.Error("Failed to get file attributes, err: ", err, ",filePath: ", filePath)
		return err
	}

	if f.Size() > GetConfig().Statistic.LogFileSize {
		err = renameStatistLog(filePath)
		if err != nil {
			Log.Error("RenameStatistLog, Failed to split file, err:", err)
			return err
		}
	}
	return nil
}

func StatisticsLog(sd *StatisticsData) error {
	// 0. Query login information

	// 1. Create a log file
	filePath, fErr := createStatistLog("")
	if fErr != nil {
		Log.Error("StatisticsLog, Failed to create log file, fErr: ", fErr)
		return fErr
	}
	// 2. Determine the file size and split large files
	splErr := splitStatistLog(filePath)
	if splErr != nil {
		Log.Error("StatisticsLog, File segmentation failed, splErr: ", splErr)
		return splErr
	}
	// 3. Convert the data format to a writable file format
	byteData := dataFormatConver(*sd)
	// 4. Write the data to a file in a fixed format
	writeErr := writeStatistLog(filePath, byteData)
	if writeErr != nil {
		Log.Error("StatisticsLog, Failed to write data, writeErr: ", writeErr)
		return writeErr
	}
	return nil
}
