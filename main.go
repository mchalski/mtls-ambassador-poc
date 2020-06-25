package main

import (
	"log"

	"github.com/mendersoftware/mtls-ambassador-poc/mender"
)

func main() {
	log.Println("reading config")
	config := ReadConfig()

	log.Printf("logging in to Mender to get mgmt token, user: %s\n", config.MenderUser)
	apiClient := mender.NewClient()
	mgmtToken, err := apiClient.Login(config.MenderUser,
		config.MenderPass,
		config.MenderBackend)

	if err != nil {
		panic(err)
	}

	log.Println("logging in to Mender: success")

	log.Println("starting server")
	h := NewHandler(apiClient, config, mgmtToken)

	s, err := NewServer(config, h)
	if err != nil {
		panic(err)
	}

	err = s.Run()
	if err != nil {
		panic(err)
	}
}
