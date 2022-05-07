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
	m.CreateTime = time.Now().In(util.CnTime)
	result := o.Create(m)
	return result.Error
}

// UpdateKickStart
func UpdateKickStart(m *KickStart) (err error) {
	o := util.GetDB()
	m.UpdateTime = time.Now().In(util.CnTime)
	result := o.Model(m).Select("content", "name", "desc", "update_time").Updates(m)
	if result.Error != nil {
		return result.Error
	}
	result = o.Find(m)

	return result.Error
}

func GetKickStartByID(id int) (*KickStart, error) {
	o := util.GetDB()
	kickStart := new(KickStart)
	kickStart.ID = id
	tx := o.Model(kickStart).Find(kickStart)
	if tx.RowsAffected == 0 {
		return nil, tx.Error
	}
	return kickStart, tx.Error
}

// GetMyKickStart
func GetMyKickStart(userid int, offset int, limit int) (total int64, ml []*KickStart, err error) {
	o := util.GetDB()
	kickStart := new(KickStart)
	tx := o.Model(kickStart).Where("user_id", userid)
	tx.Count(&total)
	tx.Limit(limit).Offset(offset).Order("id desc").Scan(&ml)
	return
}

// DeleteKickStartById
func DeleteKickStartById(userid int, id int) (deleteNum int, err error) {
	o := util.GetDB()
	m := new(KickStart)
	m.ID = id
	result := o.Debug().Model(m).Where("user_id", userid).Delete(m)
	result = o.Delete(m)
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
func preNUm(data byte) int {
	var mask byte = 0x80
	var num int = 0
	//8bit中首个0bit前有多少个1bits
	for i := 0; i < 8; i++ {
		if (data & mask) == mask {
			num++
			mask = mask >> 1
		} else {
			break
		}
	}
	return num
}

//check data is utf8
func IsUtf8(data []byte) bool {
	i := 0
	for i < len(data) {
		if (data[i] & 0x80) == 0x00 {
			i++
			continue
		} else if num := preNUm(data[i]); num > 2 {
			i++
			for j := 0; j < num-1; j++ {
				if (data[i] & 0xc0) != 0x80 {
					return false
				}
				i++
			}
		} else {
			return false
		}
	}
	return true
}
