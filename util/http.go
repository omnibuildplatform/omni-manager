package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/astaxie/beego/logs"
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
		// logs.Error("Post request failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// logs.Error("post failed, err: ", err, "body: ", requestBody)
		return nil, err
	}
	if resp.StatusCode > 300 {
		err = fmt.Errorf("Post request failed, err:%v ", string(body))
		Log.Errorln(err)
		return nil, err
	}

	// logs.Info("post successed!, body: ", string(body))
	var bodyStr map[string]interface{}
	err = json.Unmarshal(body, &bodyStr)
	if err != nil {
		logs.Error("HTTPPost Unmarshal Error:", err, string(body))
		return nil, err
	}
	// logs.Info(bodyStr)
	return bodyStr, nil
}

//HTTPGet get request
func HTTPGet(urlpath string, query url.Values) (result map[string]interface{}, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", urlpath, nil)
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	resp, err := http.DefaultClient.Do(req) //http.Get(url)
	if err != nil {
		logs.Error("HTTPGet, error: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	status, _ := strconv.Atoi(resp.Status)
	if status > 300 {
		// logs.Error("resp.Status: ", resp.Status, resp.Header)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || body == nil {
		logs.Error("err: ", err)
		return nil, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		logs.Error("HTTPGet Unmarshal,err: ", err)
		return
	}
	return
}
