package auditory

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Auditable creates data fields for recording auditory
type Auditable struct {
	CreatedBy      string     `json:"-"`
	LastModifiedBy *string    `json:"-"`
	CreateDate     time.Time  `json:"-" gorm:"autoCreateTime"`
	LastModified   *time.Time `json:"-" gorm:"autoUpdateTime"`
}

func MapFieldToColumn(fieldName string) string {
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

func GetTokenUser(c *fiber.Ctx) string {
	if tokenUser := c.Locals("tokenUser"); tokenUser != nil {
		return tokenUser.(string)
	}
	return "unknown"
}
