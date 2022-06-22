package models

import (
	"testing"
	"time"

	"github.com/omnibuildplatform/omni-manager/util"
)

func TestAddJobLog(t *testing.T) {
	util.InitConfig("../conf/app.json")
	util.InitDB()

	var item JobLog
	item.JobName = "omni-image-a6fbd713-dbf6-4243-9a7b-f2d05a8e1609"
	item.Arch = "x86_64"
	item.Release = "openEuler2109"
	item.BuildType = "installer-iso"
	item.BasePkg = ""
	item.CustomPkg = ""
	item.UserId = 111
	item.UserName = "陈其"
	item.CreateTime = time.Now().In(util.CnTime)
	item.Status = "running"
	// item.DownloadUrl = "https://repo.test.osinfra.cn/data/browse/openEuler2109/2022-04-01/openEuler-a6fbd713-dbf6-4243-9a7b-f2d05a8e1609.iso"
	item.ConfigMapName = "cmname1648800396444616"

	err := AddJobLog(&item)
	if err != nil {
		t.Errorf("first AddJobLog error:%s", err)
		return
	}
	item.Status = "finish"
	err = AddJobLog(&item)
	if err != nil {
		t.Errorf("second AddJobLog error:%s", err)
		return
	}

	UpdateJobLogStatusById(item.JobName, JOB_STATUS_FAILED)

	UpdateJobLogStatusById(item.JobName, JOB_STATUS_FAILED)

}
