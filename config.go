package main

import (
	"os"
)

type Config struct {
	MenderBackend   string
	MenderUser      string
	MenderPass      string
	MenderMgmtToken string
}

func defaultConfig() *Config {
	return &Config{

		MenderBackend:   "staging.hosted.mender.io:443",
		MenderUser:      "mtls@mender.io",
		MenderPass:      "",
		MenderMgmtToken: "",
	}
}

func ReadConfig() *Config {
	c := defaultConfig()

	backend := os.Getenv("MTLS_MENDER_BACKEND")
	if backend != "" {
		c.MenderBackend = backend
	}

	c.MenderUser = os.Getenv("MTLS_MENDER_USER")
	if c.MenderUser == "" {
		panic("provide MTLS_MENDER_USER")
	}

	c.MenderPass = os.Getenv("MTLS_MENDER_PASS")
	if c.MenderPass == "" {
		panic("provide MTLS_MENDER_PASS")
	}

	return c
}
