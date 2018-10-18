package model

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jinzhu/gorm"
)

type CommonModel struct {
	db *gorm.DB
}

var commonDb = &gorm.DB{}

func InitCommonDb(db *gorm.DB) {
	commonDb = db
}

func BeginCommit() *gorm.DB {
	return commonDb.Begin()
}

type IntKv map[int]int

func (i *IntKv) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var obj = make(map[string]int)

	err := json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	*i = make(map[int]int)
	for k, v := range obj {
		key, _ := strconv.Atoi(k)
		(*i)[key] = v
	}

	return nil
}

func (i IntKv) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for k, v := range i {
		if buf.Len() > 1 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf(`"%d":%d`, k, v))
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (i *IntKv) Scan(value interface{}) error {
	return i.UnmarshalJSON(value.([]byte))
}

func (i IntKv) Value() (driver.Value, error) {
	return i.MarshalJSON()
}
