package models

import (
	"bytes"
	"encoding/json"
	"fmt" 
	"io/ioutil"
	"net/http" 
	"strconv"
)

//HTTPPost post request
func HTTPPost(url string, requestBody string) (map[string]interface{}, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	 
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	logs.Info("HTTPPost, response Status:", resp.Status)
	logs.Info("HTTPPost, response Headers:", resp.Header)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		logs.Error("post failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	logs.Info("post successed!, body: ", string(body))
	var bodyStr map[string]interface{}
	err = json.Unmarshal(body, &bodyStr)
	if err != nil {
		logs.Error(err, string(body))
		return nil, err
	}
	logs.Info(bodyStr)
	return bodyStr, nil
}

//HTTPGitGet get request
func HTTPGitGet(url string) (col map[string]interface{}, err error) {
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("HTTPGitGet, error: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		logs.Error("resp.Status: ", resp.Status, resp.Header)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error("err: ", err)
		return nil, err
	}
	//logs.Info("url: ", url, "\n body: \n", string(body))
	err = json.Unmarshal(body, &col)
	if err != nil {
		logs.Error("HTTPGitGet,err: ", err)
		return col, err
	}
	return col, nil
}
