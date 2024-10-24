// Package utils coding=utf-8
// @Project : go-pubchem
// @Time    : 2024/1/18 16:43
// @Author  : chengxiang.luo
// @File    : excel_utils.go
// @Software: GoLand
package utils

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
)

func ReadSmilesFromExcel(xlsFiles string) ([]string, error) {
	var smilesList []string
	f, err := excelize.OpenFile(xlsFiles)
	if err != nil {
		return append(smilesList, ""), err
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			return
		}
	}()

	// Get all the rows in the Sheet1.
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return append(smilesList, ""), err
	}
	if len(rows) == 0 {
		return append(smilesList, ""), errors.New("Sheet is empty.")
	}
	// 跳过标题行
	for _, row := range rows[1:] {
		// 跳过第一列（索引从0开始）
		for _, colCell := range row[1:] {
			smilesList = append(smilesList, colCell)
		}
	}
	return smilesList, nil
}
