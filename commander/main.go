package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sivillakonski/firespotter/shared/dto"

	"github.com/gin-gonic/gin"
)

var (
	targets = []dto.Target{}
)

func main() {
	sourceFilePath := flag.String("conf", "conf.json", "Source of mocked data")
	flag.Parse()

	var err error
	targets, err = readMockedData(*sourceFilePath)
	if err != nil {
		log.Fatalf("failed to read the mock data: %s", err)
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()
	engine.Any("", serveMockedData)

	log.Printf("serving mocked data on port :8080")
	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("failed to serve mocked data: %s", err)
	}
}

func readMockedData(path string) ([]dto.Target, error) {
	rawTargets, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	targets := make([]dto.Target, 0)
	err = json.Unmarshal(rawTargets, &targets)
	if err != nil {
		return nil, err
	}

	return targets, nil
}

func serveMockedData(context *gin.Context) {
	context.JSON(http.StatusOK, targets)
}
