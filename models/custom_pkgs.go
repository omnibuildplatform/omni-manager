package models

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"omni-manager/util"
)

var CustomSigList CustomSigs

type CustomSigs struct {
	Sigs []string
}
type PkgItem struct {
	ShortName string `json:"short-name"`
}
type CustomPkg struct {
	RPMs []PkgItem `json:"rpms"`
}

//init custom package rpms list
func InitCustomPkgs() error {
	err := getSigs()
	if err != nil {
		return err
	}

	return nil
}

// get sig list
func getSigs() (err error) {
	url := util.GetConfig().BuildParam.CustomRpmAPI + "/sigs"
	util.Log.Infoln("get custom rpms from:", url)
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)

	if err != nil {
		util.Log.Errorln("1 get custom rpms error:", err)
		return
	}
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		util.Log.Errorln("2 get custom rpms error:", err)
		return
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		util.Log.Errorln("3 get custom rpms error:", err)
		return
	}
	err = json.Unmarshal(bodyData, &CustomSigList)
	if err != nil {
		util.Log.Errorln("4 get custom rpms error:", err)
		return
	}

	return
}

func GetCustomePkgList(release, arch, sig string) (customPkgList *CustomPkg, err error) {
	var resp *http.Response
	var req *http.Request
	req, err = http.NewRequest("GET", util.GetConfig().BuildParam.CustomRpmAPI+"/rpms", nil)
	if err != nil {

		return nil, err
	}
	q := req.URL.Query()
	q.Add("release", release)
	q.Add("arch", arch)
	// q.Add("sig", sig)
	q.Add("sig", "DB")
	req.URL.RawQuery = q.Encode()
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	customPkgList = new(CustomPkg)
	err = json.Unmarshal(bodyData, customPkgList)
	return

}
