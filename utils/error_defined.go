// Package utils coding=utf-8
// @Project : go-pubchem
// @Time    : 2024/5/7 15:01
// @Author  : chengxiang.luo
// @File    : error_defined.go
// @Software: GoLand
package utils

import (
	"github.com/gin-gonic/gin"
	"go-pubchem/pkg"
	"net/http"
)

type ResponseData struct {
	StatusCode int         `json:"statusCode"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data"`
}

func BadRequestErr(c *gin.Context, err error) {
	pkg.Logger.Error(err)
	c.JSON(http.StatusBadRequest, ResponseData{
		StatusCode: 400, Msg: err.Error(), Data: gin.H{},
	})
	c.Abort()
}

func OkRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, ResponseData{
		StatusCode: 200, Msg: msg, Data: gin.H{},
	})
}

func InternalRequestErr(c *gin.Context, err error) {
	pkg.Logger.Error(err)
	c.JSON(http.StatusInternalServerError, ResponseData{
		StatusCode: 500, Msg: err.Error(), Data: gin.H{},
	})
	c.Abort()
}

func OkRequestWithData(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, ResponseData{
		StatusCode: 200, Msg: msg, Data: data,
	})
}
