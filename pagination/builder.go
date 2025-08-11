package pagination

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tagAndValue struct {
	tagValue   string
	fieldValue any
}

type QueryFilter interface {
	Filter(*gorm.DB) (*gorm.DB, error)
}

func Builder[T QueryFilter](db *gorm.DB, page Page, queryFilter T) (*gorm.DB, error) {
	return filterValues(paginate(db, page), queryFilter)
}

func paginate(db *gorm.DB, p Page) *gorm.DB {
	db = db.Offset(p.Page - 1).
		Limit(p.Size).
		Order(fmt.Sprintf("%s %s", p.SortBy, p.SortOrder))
	return db
}

func filterValues(db *gorm.DB, queryFilter QueryFilter) (*gorm.DB, error) {
	var results []tagAndValue

	v := reflect.ValueOf(queryFilter)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if fmt.Sprintf("%v", value.Interface()) == "" {
			continue
		}

		tagValue := field.Tag.Get("filter")
		if tagValue == "" {
			continue
		}

		parts := strings.Split(tagValue, ";")
		filterString := parts[0]
		var valueType string

		for _, part := range parts[1:] {
			if after, ok := strings.CutPrefix(part, "type:"); ok {
				valueType = after
				break
			}
		}

		if valueType == "" {
			valueType = field.Type.Name()
		}

		var finalValue any
		var err error
		stringValue := fmt.Sprintf("%v", value.Interface())

		switch valueType {
		case "int":
			finalValue, err = strconv.Atoi(stringValue)
			if err != nil {
				return nil, fmt.Errorf("failed to parse int for field '%s': %w", field.Name, err)
			}
		case "uint":
			parsedUint64, err := strconv.ParseUint(stringValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse int for field '%s': %w", field.Name, err)
			}
			finalValue = uint(parsedUint64)
		case "float64":
			finalValue, err = strconv.ParseFloat(stringValue, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse float for field '%s': %w", field.Name, err)
			}
		case "float32":
			finalValue, err = strconv.ParseFloat(stringValue, 32)
			if err != nil {
				return nil, fmt.Errorf("failed to parse float32 for field '%s': %w", field.Name, err)
			}
		case "bool":
			finalValue, err = strconv.ParseBool(stringValue)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool for field '%s': %w", field.Name, err)
			}
		case "time.Time":
			finalValue, err = time.Parse("2006/01/02", stringValue)
			if err != nil {
				return nil, fmt.Errorf("failed to parse time for field '%s': %w", field.Name, err)
			}
		case "uuid.UUID":
			parsedUUID, err := uuid.Parse(stringValue)
			if err != nil {
				return nil, fmt.Errorf("failed to parse UUID for field '%s': %w", field.Name, err)
			}
			finalValue = parsedUUID
		default:
			finalValue = value.Interface()
		}

		results = append(results, tagAndValue{
			tagValue:   filterString,
			fieldValue: finalValue,
		})
	}

	for _, v := range results {
		db = db.Where(v.tagValue, v.fieldValue)
	}

	return db, nil
}
