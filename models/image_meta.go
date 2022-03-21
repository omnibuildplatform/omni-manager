package models

import (
	"fmt"
	"omni-manager/util"
	"time"
)

//post this body to backend
type ImageInputData struct {
	Id        int    `gorm:"primaryKey"`
	Packages  string ` description:"architecture"`
	Version   string ` description:"release openEuler Version"`
	BuildType string ` description:"iso , zip ...."`
	// BasePkg      []pkgData ` description:"default package"`
	CustomPkg []string ` description:"custom"`
}

//rpm package
type pkgData struct {
	PkgName string
	PkgMd5  string
}

type ImageMeta struct {
	Id         int       `gorm:"primaryKey"`
	Packages   string    ` description:"architecture"`
	Version    string    ` description:"release openEuler Version"`
	BuildType  string    ` description:"iso , zip ...."`
	BasePkg    string    ` description:"default package"`
	CustomPkg  string    ` description:"custom"`
	UserId     int       ` description:"user id"`
	UserName   string    ` description:"user name"`
	CreateTime time.Time ` description:"create time"`
	Status     string    ` description:"current status :running ,success, failed"`
	JobName    string    ` description:"pod name"`
}

func (t *ImageMeta) TableName() string {
	return "metadata"
}

func (t *ImageMeta) ToString() string {
	return fmt.Sprintf("id:%d;Architecture:%s;EulerVersion:%s;OutFormat:%s;UserId:%d;UserName:%s;JobName:%s", t.Id, t.Packages, t.Version, t.BuildType, t.UserId, t.UserName, t.JobName)
}

// AddMetadata insert a new Metadata into database and returns
// last inserted Id on success.
func AddMetadata(m *ImageMeta) (id int64, err error) {
	o := util.GetDB()
	result := o.Create(m)
	return int64(m.Id), result.Error
}

// GetMetadataById retrieves Metadata by Id. Returns error if
// Id doesn't exist
func GetMetadataById(id int) (v *ImageMeta, err error) {
	o := util.GetDB()
	v = &ImageMeta{Id: id}
	o.First(v, id)
	return v, err
}

// GetAllMetadata retrieves all Metadata matches certain condition. Returns empty list if
// no records exist
func GetAllMetadata(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	return nil, err
}

// UpdateMetadata updates Metadata by Id and returns error if
// the record to be updated doesn't exist
func UpdateMetadataById(m *ImageMeta) (err error) {
	o := util.GetDB()
	result := o.Model(m).Updates(m)
	return result.Error
}

// UpdateJobStatus
func UpdateJobStatus(m *ImageMeta) (err error) {
	o := util.GetDB()
	result := o.Model(m).Where("id = ?", m.Id).Update("status", m.Status)
	if result.Error == nil {
		//record log after update status
		util.Log.Infof("jobid:%d,jobname:%s, update status = %s ", m.Id, m.JobName, m.Status)
	}
	return result.Error
}

// DeleteMetadata deletes Metadata by Id and returns error if
// the record to be deleted doesn't exist
func DeleteMetadata(id int) (err error) {
	o := util.GetDB()
	temp := new(ImageMeta)
	temp.Id = id
	result := o.Delete(temp)
	return result.Error
}
