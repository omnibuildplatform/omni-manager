package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/omnibuildplatform/omni-manager/util"
)

const (
	ImageStatusStart       string = "created"
	ImageStatusDownloading string = "downloading"
	ImageStatusDone        string = "succeed"
	ImageStatusFailed      string = "failed"
)

var (
	imageStatusLock sync.Mutex
)

type BaseImagesKickStart struct {
	Label            string ` description:"name"`
	Desc             string ` description:"desc"`
	BaseImageID      string ` description:"BaseImages ID"`
	KickStartID      string ` description:"KickStart ID"`
	KickStartContent string ` description:"KickStart Content"`
	KickStartName    string ` description:"KickStart name"`
}

type BaseImages struct {
	ID         int       ` description:"id" gorm:"primaryKey"`
	Name       string    ` description:"name"  validate:"required"`
	ExtName    string    ` description:"ext name"`
	Desc       string    ` description:"desc"  validate:"required"`
	Checksum   string    ` description:"checksum"  validate:"required"`
	Algorithm  string    ` description:"algorithm"  validate:"required"`
	Url        string    ` description:"url" gorm:"-"`
	Arch       string    ` description:"arch"`
	Status     string    ` description:"status"`
	UserId     int       ` description:"user id"`
	CreateTime time.Time ` description:"create time"`
}

type ImageRequest struct {
	Name              string `description:"name"  form:"name" json:"name" validate:"required"`
	Desc              string `description:"desc"  form:"desc" json:"desc"`
	Checksum          string `description:"checksum" form:"checksum" json:"checksum"`
	Algorithm         string `description:"algorithm" form:"algorithm" json:"algorithm" validate:"required,oneof=md5 sha256"`
	ExternalID        string `description:"externalID" form:"externalID" json:"externalID" validate:"required"`
	SourceUrl         string `description:"source url of images" json:"sourceUrl" form:"sourceUrl"`
	FileName          string `description:"file name" form:"fileName" json:"fileName" validate:"required"`
	UserId            int    `description:"user id" form:"userID" json:"userID" validate:"required"`
	Publish           bool   `description:"publish image to third party storage" form:"publish" json:"publish"  `
	ExternalComponent string `description:"From APP" form:"externalComponent" json:"externalComponent" validate:"required"`
}
type ImageResponse struct {
	ImageRequest
	ID           int       `description:"id" form:"id" json:"id"`
	Status       string    `description:"image status" json:"status"`
	StatusDetail string    `description:"status detail"  json:"statusDetail"`
	ImagePath    string    `description:"image store path"  json:"imagePath"`
	ChecksumPath string    `description:"image checksum store path"  json:"checksumPath"`
	CreateTime   time.Time `description:"create time" json:"createTime"`
	UpdateTime   time.Time `description:"update time" json:"updateTime"`
}

func (t *BaseImages) TableName() string {
	return "base_images"
}

type ImageStatusEvent struct {
	BlockSize int `json:"blockSize"`
	ImageSize int `json:"imageSize"`
}

// AddBaseImages insert a new BaseImages into database and returns
// last inserted Id on success.
func AddBaseImages(m *BaseImages) (err error) {
	o := util.GetDB()
	result := o.Create(m)
	return result.Error
}

func GetBaseImagesByID(id int) (v *BaseImages, err error) {
	o := util.GetDB()
	v = new(BaseImages)
	v.ID = id
	tx := o.Model(v).Find(v)
	if tx.RowsAffected == 0 {
		return nil, tx.Error
	}
	return v, tx.Error
}

// GetMyBaseImages
func GetMyBaseImages(userid int, offset int, limit int) (total int64, ml []*BaseImages, err error) {
	o := util.GetDB()
	baseImages := new(BaseImages)
	tx := o.Model(baseImages).Where("user_id", userid)
	tx.Count(&total)
	tx.Limit(limit).Offset(offset).Order("id desc").Scan(&ml)
	return
}

// DeleteBaseImagesById
func DeleteBaseImagesById(userid, id int) (deleteNum int, err error) {
	o := util.GetDB()
	m := new(BaseImages)
	m.ID = id
	result := o.Model(m).Where("user_id", userid).Delete(m)
	return int(result.RowsAffected), result.Error
}

// UpdateBaseImages
func UpdateBaseImages(m *BaseImages) (err error) {
	o := util.GetDB()
	result := o.Model(m).Select("checksum", "name", "desc", "url", "arch", "status").Updates(m)
	if result.Error != nil {
		return result.Error
	}
	result = o.Find(m)
	return result.Error
}

// UpdateBaseImagesStatus
func UpdateBaseImagesStatus(m *BaseImages) (err error) {
	o := util.GetDB()
	result := o.Debug().Model(m).Select("status").Updates(m)
	return result.Error
}

func handleDownloadStatusEvent(ctx context.Context, event cloudevents.Event) {
	var imageStatusEvent ImageStatusEvent
	err := json.Unmarshal(event.Data(), &imageStatusEvent)
	if err != nil {
		log.Printf(" handleDownloadStatusEvent error : %v     \n", err)
		return
	}
	externalItems := strings.Split(event.Subject(), ".")
	if len(externalItems) < 2 {
		log.Printf("cloudeEvent's subject is not format:  externalComponent.externalID     . it's: %s  \n", event.Subject())
		return
	}
	// check event subject is from mine or not
	if externalItems[0] != util.GetConfig().AppName {
		return
	}

	key := fmt.Sprintf("imageStatus:%s:%s", externalItems[0], externalItems[1])

	switch event.Type() {
	case string(ImageDownloaded):
		//for parallel
		imageStatusLock.Lock()
		finifshSize, _ := util.GetFloat(key)

		tempInt := math.Ceil(finifshSize)
		tempInt += float64(imageStatusEvent.BlockSize)
		finifshSize = tempInt + (tempInt)/float64(imageStatusEvent.ImageSize)
		util.Set(key, finifshSize)
		imageStatusLock.Unlock()
	case string(ImageFailed):
		var baseimage BaseImages
		baseimage.ID, _ = strconv.Atoi(externalItems[1])
		baseimage.Status = ImageStatusFailed
		UpdateBaseImagesStatus(&baseimage)
		//if download failed 。delete item
		util.DelKey(key, nil)
	case string(ImageVerified):
		var baseimage BaseImages
		baseimage.ID, _ = strconv.Atoi(externalItems[1])
		baseimage.Status = ImageStatusDone
		UpdateBaseImagesStatus(&baseimage)
		util.DelKey(key, nil)
	default:

	}

}
