package main

import (
	_ "github.com/go-sql-driver/mysql"
	"go-pubchem/pkg"
	"go-pubchem/router"
	"math/rand"
	"time"
)

// var parsedNames []string
// var unparsedNames []string
func randomInt(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())

	if min >= max {
		return min
	}

	rangeFloat := float64(max-min) + 1
	randomFloat := rand.Float64() * rangeFloat

	return int64(randomFloat) + min
}

// @title Golang go-pubchem APIs
// @version 1.0
// @description this is a sample server celler server
// @termsOfService https://www.swagger.io/terms/

// @contact.name chengxiangLuo
// @contact.url https://github.com/netluo
// @contact.email andrew.luo1992@gmile.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:8100
// @BasePath /api/v1
func main() {
	r := router.NewRouter("./go-pubchem.log", "INFO")
	pkg.Logger.Info(r.Run("0.0.0.0:8100"))
}
