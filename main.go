package main

import (
	"log"
	"os"

	"github.com/mendersoftware/mtls-ambassador/mender"
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {
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
