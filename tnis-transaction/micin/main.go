package main

import (
	"fmt"
	"tnis/tnis-transaction/micin/app"
	"tnis/tnis-transaction/micin/config"
)

func main() {
	config := config.GetConfig()

	port := config.SETTING.Port

	app := &app.App{}
	app.Initialize(config)
	fmt.Println("running on port ", port)
	app.Run(port)
}
