package main

import (
	"fmt"
	"github.com/grd888/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	cfg.SetUser("grd888")
	cfg, err = config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	fmt.Println("Config:", cfg)
}