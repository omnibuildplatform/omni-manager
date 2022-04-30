package models

import (
	"fmt"
	"time"

	"github.com/omnibuildplatform/omni-manager/util"
)

type BaseImagesKickStart struct {
	BaseImageID      int    ` description:"id"`
	KickStartContent string ` description:"KickStart Content"`
}

type BaseImages struct {
	ID         int       ` description:"id" gorm:"primaryKey"`
	Name       string    ` description:"name"`
	Desc       string    ` description:"desc"`
	checksum   string    ` description:"checksum"`
	Url        string    ` description:"url"`
	Arch       string    ` description:"arch"`
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
	result := o.FirstOrCreate(m)
	return result.Error
}

func GetBaseImagesByJobName(jobname string) (v *BaseImages, err error) {
	o := util.GetDB()
	v = new(BaseImages)
	sql := fmt.Sprintf("select * from %s where job_name = '%s' order by create_time desc limit 1", v.TableName(), jobname)
	tx := o.Raw(sql).Scan(v)
	return v, tx.Error
}

// GetAllBaseImages retrieves all BaseImages matches certain condition. Returns empty list if
// no records exist
func GetAllBaseImages(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	return nil, err
}

// DeleteBaseImagesById
func DeleteBaseImagesById(userid, id int) (err error) {
	o := util.GetDB()
	m := new(BaseImages)
	m.ID = id
	m.UserId = userid
	result := o.Delete(m)
	return result.Error
}

// UpdateBaseImages
func UpdateBaseImages(m *BaseImages) (err error) {
	o := util.GetDB()
	result := o.Updates(m)
	return result.Error
}
