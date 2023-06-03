package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("./config.yaml")
	fmt.Println("Hello, World!")
}
