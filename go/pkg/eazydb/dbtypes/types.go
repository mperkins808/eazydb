package dbtypes

import (
	"fmt"
	"reflect"
	"time"
)

type ValType string

const (
	TEXT     ValType = "TEXT"
	INT      ValType = "INT"
	FLOAT    ValType = "FLOAT"
	DOUBLE   ValType = "DOUBLE"
	DATETIME ValType = "DATETIME"
	BOOL     ValType = "BOOL"
	BLOB     ValType = "BLOB"
	SERIAL   ValType = "SERIAL"
	NONE     ValType = "NONE"
)

func ToSQL(gotype reflect.Kind) (ValType, error) {

	switch gotype {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return INT, nil
	case reflect.String:
		return TEXT, nil
	case reflect.Float32, reflect.Float64:
		return DOUBLE, nil
	case reflect.Struct:
		if gotype == reflect.TypeOf(time.Time{}).Kind() {
			return DATETIME, nil
		}
	default:
		return NONE, fmt.Errorf("%v is support or not yet supported", gotype.String())
	}
	return NONE, fmt.Errorf("%v is support or not yet supported", gotype.String())
}

type Key struct {
	Name    string
	SQLType ValType
}
