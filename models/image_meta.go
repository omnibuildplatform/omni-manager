package models

import (
	"fmt"
	"omni-manager/util"
	"time"
)

//post this body to backend
type ImageInputData struct {
	Id        int      `gorm:"primaryKey"`
	Arch      string   ` description:"architecture"`
	Release   string   ` description:"release openEuler Version"`
	BuildType string   ` description:"iso , zip ...."`
	CustomPkg []string ` description:"custom"`
}

//rpm package
type pkgData struct {
	PkgName string
	PkgMd5  string
}

type ImageMeta struct {
	Id            int       `gorm:"primaryKey"`
	Arch          string    ` description:"architecture"`
	Release       string    ` description:"release openEuler Version"`
	BuildType     string    ` description:"iso , zip ...."`
	BasePkg       string    ` description:"default package"`
	CustomPkg     string    ` description:"custom"`
	UserId        int       ` description:"user id"`
	UserName      string    ` description:"user name"`
	CreateTime    time.Time ` description:"create time"`
	Status        string    ` description:"current status :running ,success, failed"`
	JobName       string    ` description:"pod name"`
	DownloadUrl   string    ` description:"download the result of build iso file"`
	ConfigMapName string    ` description:"configMap name"`
}

func (t *ImageMeta) TableName() string {
	return "image_meta"
}

func (t *ImageMeta) ToString() string {
	return fmt.Sprintf("id:%d;Architecture:%s;EulerVersion:%s;OutFormat:%s;UserId:%d;UserName:%s;JobName:%s", t.Id, t.Arch, t.Release, t.BuildType, t.UserId, t.UserName, t.JobName)
}

// AddImageMeta insert a new ImageMeta into database and returns
// last inserted Id on success.
func AddImageMeta(m *ImageMeta) (id int64, err error) {
	o := util.GetDB()
	result := o.Create(m)
	return int64(m.Id), result.Error
}

// GetImageMetaById retrieves ImageMeta by Id. Returns error if
// Id doesn't exist
func GetImageMetaById(id int) (v *ImageMeta, err error) {
	o := util.GetDB()
	v = &ImageMeta{Id: id}
	o.First(v, id)
	return v, err
}

// GetAllImageMeta retrieves all ImageMeta matches certain condition. Returns empty list if
// no records exist
func GetAllImageMeta(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	return nil, err
}

// GetMyImageMetaHistory query my build history
func GetMyImageMetaHistory(userid int, offset int, limit int) (ml []*ImageMeta, err error) {
	o := util.GetDB()
	m := new(ImageMeta)
	m.UserId = userid
	ml = make([]*ImageMeta, limit)
	sql := fmt.Sprintf("select * from %s where user_id = %d order by id desc limit %d,%d", m.TableName(), userid, offset, limit)
	o.Raw(sql).Scan(&ml)
	return ml, nil
}

// UpdateImageMeta updates ImageMeta by Id and returns error if
// the record to be updated doesn't exist
func UpdateImageMetaById(m *ImageMeta) (err error) {
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

// CreateTables
func CreateTables() (err error) {
	o := util.GetDB()
	err = o.Migrator().CreateTable(&ImageMeta{})
	return
}
