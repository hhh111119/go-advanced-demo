package homework

import (
	"errors"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")

func InsertStmt(entity interface{}) (string, []interface{}, error) {
	if entity == nil {
		return "", nil, errInvalidEntity
	}
	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Pointer {
		return insertStmt(val.Elem().Interface())
	}
	return insertStmt(entity)
}

func insertStmt(entity interface{}) (string, []interface{}, error) {

	if entity == nil {
		return "", nil, errInvalidEntity
	}

	val := reflect.ValueOf(entity)
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Struct {
		return "", nil, errInvalidEntity
	}
	tableName := getTableName(typ)
	colValues := make([]interface{}, 0, typ.NumField())
	cols := make([]string, 0, typ.NumField())
	if err := parseColumns(typ, val, &colValues, &cols); err != nil {
		return "", nil, err
	}
	if len(colValues) == 0 {
		return "", nil, errInvalidEntity
	}

	return genInsertSQL(tableName, cols), colValues, nil

}

func genInsertSQL(tableName string, cols []string) string {
	bd := strings.Builder{}
	bd.WriteString("INSERT INTO ")
	bd.WriteString(quoteString(tableName))
	bd.WriteString("(")
	valuesBD := strings.Builder{}
	valuesBD.WriteString("VALUES(")
	for id, col := range cols {
		bd.WriteString(quoteString(col))
		valuesBD.WriteString("?")
		if id != len(cols)-1 {
			bd.WriteString(",")
			valuesBD.WriteString(",")
		}
	}
	bd.WriteString(") ")
	valuesBD.WriteString(")")
	bd.WriteString(valuesBD.String())
	bd.WriteString(";")
	return bd.String()
}

func parseColumns(typ reflect.Type, val reflect.Value, colValues *[]interface{}, cols *[]string) error {
	if typ.Kind() != reflect.Struct {
		return errInvalidEntity
	}
	for i := 0; i < typ.NumField(); i += 1 {
		colType := typ.Field(i)
		if !colType.IsExported() {
			continue
		}
		colValue := val.Field(i)
		if colValue.Kind() == reflect.Struct && colType.Anonymous {
			if err := parseColumns(colType.Type, colValue, colValues, cols); err != nil {
				return err
			}
			continue
		}
		exist := false
		for _, existName := range *cols {
			if existName == colType.Name {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		*cols = append(*cols, colType.Name)
		*colValues = append(*colValues, colValue.Interface())
	}
	return nil
}

func getTableName(typ reflect.Type) string {
	return typ.Name()
}

func quoteString(s string) string {
	return "`" + s + "`"
}
