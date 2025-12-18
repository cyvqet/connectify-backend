package main

import (
	"flag"
	"os"

	"github.com/spf13/viper"
)

func main() {
	initConfig() // initialize config
	server := InitWebServer()
	server.Run(":8080")
}

func initConfig() {
	// method 1: get environment type from command line parameter
	envFlag := flag.String("env", "", "environment: dev or k8s (default: dev)")
	flag.Parse()

	var env string
	if *envFlag != "" {
		env = *envFlag
	} else {
		// method 2: get environment type from environment variable
		env = os.Getenv("ENV")
		if env == "" {
			env = "dev" // default use dev environment
		}
	}

	viper.SetConfigName(env)      // config file name (dev or k8s)
	viper.SetConfigType("yaml")   // config file type
	viper.AddConfigPath("config") // config file path
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
