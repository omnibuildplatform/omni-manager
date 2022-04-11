package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"omni-manager/util"
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
		util.Log.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	util.Log.Info("HTTPPost, response Status:", resp.Status)
	util.Log.Info("HTTPPost, response Headers:", resp.Header)
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		util.Log.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	if err != nil || body == nil {
		util.Log.Error("post failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	util.Log.Info("post successed!, body: ", string(body))
	var bodyStr map[string]interface{}
	err = json.Unmarshal(body, &bodyStr)
	if err != nil {
		util.Log.Error(err, string(body))
		return nil, err
	}
	util.Log.Info(bodyStr)
	return bodyStr, nil
}

//HTTPGitGet get request
func HTTPGitGet(url string) (col map[string]interface{}, err error) {
	resp, err := http.Get(url)
	if err != nil {
		util.Log.Error("HTTPGitGet, error: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		util.Log.Error("resp.Status: ", resp.Status, resp.Header)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		util.Log.Error("err: ", err)
		return nil, err
	}
	//util.Log.Info("url: ", url, "\n body: \n", string(body))
	err = json.Unmarshal(body, &col)
	if err != nil {
		util.Log.Error("HTTPGitGet,err: ", err)
		return col, err
	}
	return col, nil
}
