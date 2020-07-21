package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/urfave/cli"

	api "github.com/mendersoftware/mtls-ambassador/api/http"
	"github.com/mendersoftware/mtls-ambassador/app"
	"github.com/mendersoftware/mtls-ambassador/client/mender"
	aconfig "github.com/mendersoftware/mtls-ambassador/config"
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {
	var configPath string

	app := &cli.App{
		Name: "mtls-ambassador",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Configuration `FILE`. Supports JSON, TOML, YAML and HCL formatted configs.",
				Value:       "config.yaml",
				Destination: &configPath,
			},
		},
		Action: cmdServer,
	}

	app.Before = func(args *cli.Context) error {
		err := config.FromConfigFile(configPath, aconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		config.Config.SetEnvPrefix("MTLS")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

func cmdServer(args *cli.Context) error {
	if err := validateConfig(config.Config); err != nil {
		log.Fatal(err)
	}

	backend := config.Config.GetString(
		aconfig.SettingMenderBackend,
	)

	proxy, err := api.NewProxy(backend)
	if err != nil {
		log.Fatal(err)
	}

	client := mender.NewClient(backend)

	user := config.Config.GetString(
		aconfig.SettingMenderUser,
	)
	pass := config.Config.GetString(
		aconfig.SettingMenderPass,
	)

	authProvider, err := app.NewAuthProvider(client, user, pass)
	if err != nil {
		log.Fatal(err)
	}

	app := app.NewApp(client, authProvider)
	r, err := api.NewRouter(app, proxy)
	if err != nil {
		log.Fatal(err)
	}

	srvCertFile := config.Config.GetString(
		aconfig.SettingServerCert,
	)
	srvKeyFile := config.Config.GetString(
		aconfig.SettingServerKey,
	)
	tenantCACertFile := config.Config.GetString(
		aconfig.SettingTenantCAPem,
	)
	port := config.Config.GetString(
		aconfig.SettingListen,
	)

	s, err := NewServer(r,
		srvCertFile,
		srvKeyFile,
		tenantCACertFile,
		port)
	if err != nil {
		log.Fatal(err)
	}

	return s.Run()
}

func validateConfig(c config.Reader) error {
	required := []string{
		aconfig.SettingMenderBackend,
		aconfig.SettingMenderUser,
		aconfig.SettingMenderPass,
	}

	for _, c := range required {
		if config.Config.GetString(c) == "" {
			return errors.New(fmt.Sprintf("provide setting %s\n", c))
		}
	}

	return nil
}
