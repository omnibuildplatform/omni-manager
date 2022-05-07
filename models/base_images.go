package models

import (
	"fmt"
	"time"

	"github.com/omnibuildplatform/omni-manager/util"
)

type BaseImagesKickStart struct {
	Name             string ` description:"name"`
	Desc             string ` description:"desc"`
	BaseImageID      string ` description:"BaseImages ID"`
	KickStartContent string ` description:"KickStart Content"`
	KickStartName    string ` description:"KickStart name"`
}

type BaseImages struct {
	ID         int       ` description:"id" gorm:"primaryKey"`
	Name       string    ` description:"name"`
	Desc       string    ` description:"desc"`
	Checksum   string    ` description:"checksum"`
	Url        string    ` description:"url"`
	Arch       string    ` description:"arch"`
	Status     string    ` description:"status"`
	UserId     int       ` description:"user id"`
	CreateTime time.Time ` description:"create time"`
}

func (t *BaseImages) TableName() string {
	return "base_images"
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
	sql := fmt.Sprintf("select * from %s where id = %d ", v.TableName(), id)
	tx := o.Raw(sql).Scan(v)
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
	result := o.Debug().Model(m).Where("user_id", userid).Delete(m)
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
