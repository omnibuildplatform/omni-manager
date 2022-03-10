package models

import (
	"fmt"
	"omni-manager/util"
)

type Metadata struct {
	Id          int    `gorm:"primaryKey"`
	ProjectName string ` description:"项目名称"`
	PackageName string ` description:"包的名称"`
	ArcName     string ` description:"架构的名称。"`
	Desc        string ` description:"简介"`
}

func (t *Metadata) ToString() string {

	return fmt.Sprintf("id:%d;ProjectName:%s;PackageName:%s;ArcName:%s;Desc:%s;", t.Id, t.ProjectName, t.PackageName, t.ArcName, t.Desc)
}

// AddMetadata insert a new Metadata into database and returns
// last inserted Id on success.
func AddMetadata(m *Metadata) (id int64, err error) {
	o := util.GetDB()
	result := o.Create(m)
	return int64(m.Id), result.Error
}

// GetMetadataById retrieves Metadata by Id. Returns error if
// Id doesn't exist
func GetMetadataById(id int) (v *Metadata, err error) {
	o := util.GetDB()
	v = &Metadata{Id: id}
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
func UpdateMetadataById(m *Metadata) (err error) {
	o := util.GetDB()
	result := o.Model(m).Updates(m)
	return result.Error
}

// DeleteMetadata deletes Metadata by Id and returns error if
// the record to be deleted doesn't exist
func DeleteMetadata(id int) (err error) {
	o := util.GetDB()
	temp := new(Metadata)
	temp.Id = id
	result := o.Delete(temp)
	return result.Error
}
