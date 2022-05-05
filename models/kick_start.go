package models

import (
	"fmt"
	"time"

	"github.com/omnibuildplatform/omni-manager/util"
)

type KickStart struct {
	ID         int       ` description:"id" gorm:"primaryKey"`
	Name       string    ` description:"name"`
	Desc       string    ` description:"desc"`
	Content    string    ` description:"content"`
	UserId     int       ` description:"user id"`
	CreateTime time.Time ` description:"create time"`
	UpdateTime time.Time ` description:"update  time"`
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

// DeleteKickStartById
func DeleteKickStartById(id int) (err error) {
	o := util.GetDB()
	m := new(KickStart)
	m.ID = id
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
