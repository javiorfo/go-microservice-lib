package auditory

import (
	"time"
)

type Auditable struct {
	CreatedBy        string     `gorm:"not null" bson:"createdBy"`
	LastModifiedBy   *string    `bson:"lastModifiedBy"`
	CreateDate       time.Time  `gorm:"autoCreateTime;not null" bson:"createDate"`
	LastModifiedDate *time.Time `gorm:"autoUpdateTime" bson:"lastModifiedDate"`
}

func (Auditable) MapFieldToSQLColumn(fieldName string) string {
	fieldToColumnMap := map[string]string{
		"CreatedBy":      "created_by",
		"CreateDate":     "create_date",
		"LastModifiedBy": "last_modified_by",
		"LastModified":   "last_modified_date",
	}

	if columnName, ok := fieldToColumnMap[fieldName]; ok {
		return columnName
	}
	return fieldName
}

func (a *Auditable) Update(auditor *string) {
	a.LastModifiedBy = auditor
}

func New(auditor string) Auditable {
	return Auditable{CreatedBy: auditor}
}
