// Package utils coding=utf-8
// @Project : go-pubchem
// @Time    : 2023/10/25 16:28
// @Author  : chengxiang.luo
// @Email   : chengxiang.luo1992@gmail.com
// @File    : other_tools.go
// @Software: GoLand
package utils

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetInterfaceToString(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case time.Time:
		t, _ := value.(time.Time)
		key = t.String()
		// 2022-11-23 11:29:07 +0800 CST  这类格式把尾巴去掉
		key = strings.Replace(key, " +0800 CST", "", 1)
		key = strings.Replace(key, " +0000 UTC", "", 1)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func ReadRecordsFromCsv(csvFilePath string) [][]string {
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		panic(err)
	}
	defer func(csvFile *os.File) {
		err := csvFile.Close()
		if err != nil {
			panic(err)
		}
	}(csvFile)

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	return records
}

func GetInterfaceToInt(t1 interface{}) int {
	var t2 int
	switch t1.(type) {
	case uint:
		t2 = int(t1.(uint))
		break
	case int8:
		t2 = int(t1.(int8))
		break
	case uint8:
		t2 = int(t1.(uint8))
		break
	case int16:
		t2 = int(t1.(int16))
		break
	case uint16:
		t2 = int(t1.(uint16))
		break
	case int32:
		t2 = int(t1.(int32))
		break
	case uint32:
		t2 = int(t1.(uint32))
		break
	case int64:
		t2 = int(t1.(int64))
		break
	case uint64:
		t2 = int(t1.(uint64))
		break
	case float32:
		t2 = int(t1.(float32))
		break
	case float64:
		t2 = int(t1.(float64))
		break
	case string:
		t2, _ = strconv.Atoi(t1.(string))
		if t2 == 0 && len(t1.(string)) > 0 {
			f, _ := strconv.ParseFloat(t1.(string), 64)
			t2 = int(f)
		}
		break
	case nil:
		t2 = 0
		break
	case json.Number:
		t3, _ := t1.(json.Number).Int64()
		t2 = int(t3)
		break
	default:
		t2 = t1.(int)
		break
	}
	return t2
}
