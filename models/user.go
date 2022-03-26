package models

import (
	"omni-manager/util"

	"github.com/Authing/authing-go-sdk/lib/management"
)

func test() {
	client := management.NewClient(util.GetConfig().AppID, util.GetConfig().AppSecret)
	resp, err := client.ExportAll()
	_, _ = resp, err
	client.UserPoolDetail()
}
