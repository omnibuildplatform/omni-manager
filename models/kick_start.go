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

// UpdateKickStart
func UpdateKickStart(m *KickStart) (err error) {
	o := util.GetDB()
	result := o.Updates(m)
	return result.Error
}

// GetMyKickStart
func GetMyKickStart(userid int, offset int, limit int) (total int64, ml []*KickStart, err error) {
	o := util.GetDB()
	kickStart := new(KickStart)
	kickStart.UserId = userid
	tx := o.Model(kickStart)
	tx.Count(&total)
	tx.Limit(limit).Offset(offset).Order("id desc").Scan(&ml)
	return
}

// DeleteKickStartById
func DeleteKickStartById(userid int, id int) (deleteNum int, err error) {
	o := util.GetDB()
	m := new(KickStart)
	m.ID = id
	m.UserId = userid
	result := o.Delete(m)
	deleteNum = int(result.RowsAffected)
	err = result.Error
	return
}

// DeleteMultiKickStarts
func DeleteMultiKickStarts(names string) (err error) {
	o := util.GetDB()
	m := new(KickStart)
	sql := fmt.Sprintf("delete from %s  where job_name in (%s)", m.TableName(), names)
	result := o.Model(m).Exec(sql)
	return result.Error
}
