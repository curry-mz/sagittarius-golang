package utils

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// GenBatchUpdateSQLWithPK dataList中需要包含主键, updateColumn为需要更新的字段
func GenBatchUpdateSQLWithPK(tableName string, dataList interface{}, pk string, updateColumn []string) ([]string, error) {
	fieldValue := reflect.ValueOf(dataList)
	fieldType := reflect.TypeOf(dataList).Elem().Elem()
	sliceLength := fieldValue.Len()
	fieldNum := fieldType.NumField()

	var IDList []string
	updateMap := make(map[string][]string)
	for i := 0; i < sliceLength; i++ {
		structValue := fieldValue.Index(i).Elem()
		for j := 0; j < fieldNum; j++ {
			elem := structValue.Field(j)

			var tid string
			switch elem.Kind() {
			case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int8, reflect.Int16:
				tid = strconv.FormatInt(elem.Int(), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				tid = strconv.FormatUint(elem.Uint(), 10)
			case reflect.String:
				if strings.Contains(elem.String(), "'") {
					tid = fmt.Sprintf("'%v'", strings.ReplaceAll(elem.String(), "'", "\\'"))
				} else {
					tid = fmt.Sprintf("'%v'", elem.String())
				}
			case reflect.Float64, reflect.Float32:
				tid = strconv.FormatFloat(elem.Float(), 'f', -1, 64)
			case reflect.Bool:
				tid = strconv.FormatBool(elem.Bool())
			default:
				return nil, fmt.Errorf("type conversion error, param is %v", fieldType.Field(j).Tag.Get("json"))
			}

			gormTag := fieldType.Field(j).Tag.Get("gorm")
			fieldTag := getFieldName(gormTag)

			if strings.HasPrefix(fieldTag, pk) {
				id, err := strconv.ParseInt(tid, 10, 64)
				if err != nil {
					return nil, err
				}

				if id < 1 {
					return nil, fmt.Errorf("this structure should have a primary key and gt 0")
				}
				IDList = append(IDList, tid)
				continue
			}
			needUpdate := false
			for _, column := range updateColumn {
				if fieldTag == column {
					needUpdate = true
					break
				}
			}
			if needUpdate {
				valueList := append(updateMap[fieldTag], tid)
				updateMap[fieldTag] = valueList
			}
		}
	}
	if len(updateMap) == 0 {
		return nil, nil
	}

	length := len(IDList)
	// Length of each batch submission
	size := 2000
	SQLQuantity := int(math.Ceil(float64(length) / float64(size)))

	var SQLArray []string
	k := 0

	for i := 0; i < SQLQuantity; i++ {
		count := 0
		var record bytes.Buffer
		record.WriteString("UPDATE " + tableName + " SET ")

		for fieldName, fieldValueList := range updateMap {
			record.WriteString(fieldName)
			record.WriteString(" = CASE " + pk)

			for j := k; j < len(IDList) && j < len(fieldValueList) && j < size+k; j++ {
				record.WriteString(" WHEN " + IDList[j] + " THEN " + fieldValueList[j])
			}

			count++
			if count != len(updateMap) {
				record.WriteString(" END, ")
			}
		}
		record.WriteString(" END WHERE ")
		record.WriteString(pk + " IN (")
		min := size + k
		if len(IDList) < min {
			min = len(IDList)
		}
		record.WriteString(strings.Join(IDList[k:min], ","))
		record.WriteString(");")

		k += size
		SQLArray = append(SQLArray, record.String())
	}
	return SQLArray, nil
}

func getFieldName(fieldTag string) string {
	fieldTagArr := strings.Split(fieldTag, ":")
	if len(fieldTagArr) == 0 {
		return ""
	}

	return fieldTagArr[len(fieldTagArr)-1]
}
