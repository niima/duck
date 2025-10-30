package main

import (
	"duck/common"
	"duck/httputils"
	"fmt"
)

func main() {
	logger := common.NewLogger("app3")
	config := common.NewConfig("app3")

	logger.Info(fmt.Sprintf("Starting %s on port %s", config.AppName, config.Port))

	rw := httputils.NewResponseWriter()
	logger.Info(fmt.Sprintf("Response writer initialized: %v", rw))

	fmt.Println("Hello, World! from user-api (app3)")
}
