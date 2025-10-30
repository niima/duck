package main

import (
	"duck/common"
	"duck/httputils"
	"fmt"
)

func main() {
	logger := common.NewLogger("app1")
	config := common.NewConfig("app1")

	logger.Info(fmt.Sprintf("Starting %s on port %s", config.AppName, config.Port))

	client := httputils.NewClient()
	logger.Info(fmt.Sprintf("HTTP client initialized: %v", client))

	fmt.Println("Hello, World! from event-service (app1)")
}
