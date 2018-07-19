package main

import (
	"fmt"
	"tnis/tnis-gateway/micin/app"
	"tnis/tnis-gateway/micin/config"
)

func main() {
	config := config.GetConfig()

	port := config.SETTING.Port

	app := &app.App{}
	app.Initialize(config)
	fmt.Println("running on port ", port)
	app.Run(port)
}
