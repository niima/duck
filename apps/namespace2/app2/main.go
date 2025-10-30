package main

import (
	"duck/common"
	"fmt"
)

func main() {
	logger := common.NewLogger("app2")
	config := common.NewConfig("app2")

	logger.Info(fmt.Sprintf("Starting %s on port %s", config.AppName, config.Port))

	fmt.Println("Hello, World! from profile-service (app2)")
}
