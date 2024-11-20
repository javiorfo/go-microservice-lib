package auditory

import (
	"time"
)

// @Model Auditable
// @Description creates data fields for recording auditory
// @ID auditory.Auditable
type Auditable struct {
	CreatedBy      string     `json:"-" bson:"createdBy"`
	LastModifiedBy *string    `json:"-" bson:"lastModifiedBy"`
	CreateDate     time.Time  `json:"-" gorm:"autoCreateTime" bson:"createDate"`
	LastModified   *time.Time `json:"-" gorm:"autoUpdateTime" bson:"lastModified"`
}

func (Auditable) MapFieldToSQLColumn(fieldName string) string {
	fieldToColumnMap := map[string]string{
		"CreatedBy":      "created_by",
		"CreateDate":     "create_date",
		"LastModifiedBy": "last_modified_by",
		"LastModified":   "last_modified",
	}

	if columnName, ok := fieldToColumnMap[fieldName]; ok {
		return columnName
	}
	return fieldName
}
