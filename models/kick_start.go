package models

import (
	"fmt"
	"time"

	"github.com/omnibuildplatform/omni-manager/util"
)

type KickStart struct {
	JobName       string    ` description:"pod name" gorm:"primaryKey"`
	Arch          string    ` description:"architecture"`
	Release       string    ` description:"release openEuler Version"`
	BuildType     string    ` description:"iso , zip ...."`
	BasePkg       string    ` gorm:"size:5055"  description:"default package"`
	CustomPkg     string    ` gorm:"size:5055" description:"custom"`
	UserId        int       ` description:"user id"`
	UserName      string    ` description:"user name"`
	CreateTime    time.Time ` description:"create time"`
	Status        string    ` description:"current status :running ,success, failed"`
	DownloadUrl   string    ` description:"download the result of build iso file"`
	ConfigMapName string    ` description:"configMap name"`
	JobLabel      string    ` description:"job label"`
	JobDesc       string    ` description:"job description"`
	StartTime     time.Time ` description:"create time"`
	EndTime       time.Time ` description:"create time"`
}

func (t *KickStart) TableName() string {
	return "kick_start"
}

// AddKickStart insert a new ImageMeta into database and returns
// last inserted Id on success.
func AddKickStart(m *KickStart) (err error) {
	o := util.GetDB()
	result := o.FirstOrCreate(m)
	return result.Error
}

func GetKickStartByJobName(jobname string) (v *KickStart, err error) {
	o := util.GetDB()
	v = new(KickStart)
	sql := fmt.Sprintf("select * from %s where job_name = '%s' order by create_time desc limit 1", v.TableName(), jobname)
	tx := o.Raw(sql).Scan(v)
	return v, tx.Error
}

// GetAllKickStart retrieves all ImageMeta matches certain condition. Returns empty list if
// no records exist
func GetAllKickStart(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	return nil, err
}

// GetMyKickStarts query my build history
func GetMyKickStarts(jobitem *KickStart, nameOrDesc string, offset int, limit int) (total int64, ml []*KickStart, err error) {
	o := util.GetDB()
	tx := o.Model(jobitem)
	if len(jobitem.Arch) > 0 {
		tx = tx.Where("arch = ?", jobitem.Arch)
	}
	if len(jobitem.Status) > 0 {
		tx = tx.Where("status = ?", jobitem.Status)
	}
	if len(jobitem.BuildType) > 0 {
		tx = tx.Where("build_type = ?", jobitem.BuildType)
	}

	tx = tx.Where("user_id = ?", jobitem.UserId)

	if len(nameOrDesc) > 0 {
		tx = tx.Where("job_label like '%" + nameOrDesc + "%'  or job_desc like '%" + nameOrDesc + "%' ")
	}
	tx.Count(&total)
	tx.Limit(limit).Offset(offset).Order("create_time desc").Scan(&ml)
	return total, ml, tx.Error
}

// DeleteKickStartById
func DeleteKickStartById(jobName string) (err error) {
	o := util.GetDB()
	m := new(KickStart)
	m.JobName = jobName
	result := o.Delete(m)
	return result.Error
}

// DeleteMultiKickStarts
func DeleteMultiKickStarts(names string) (err error) {
	o := util.GetDB()
	m := new(KickStart)
	sql := fmt.Sprintf("delete from %s  where job_name in (%s)", m.TableName(), names)
	result := o.Model(m).Exec(sql)
	return result.Error
}
