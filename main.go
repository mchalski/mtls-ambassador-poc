// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/mendersoftware/go-lib-micro/log"
	"github.com/urfave/cli"

	api "github.com/mendersoftware/mtls-ambassador/api/http"
	"github.com/mendersoftware/mtls-ambassador/app"
	"github.com/mendersoftware/mtls-ambassador/client/mender"
	aconfig "github.com/mendersoftware/mtls-ambassador/config"
)

var (
	l = log.NewEmpty()
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {
	var configPath string

	l.Info("starting mtls-ambassador")

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
		l.Infof("loading config %s", configPath)
		err := config.FromConfigFile(configPath, aconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		config.Config.SetEnvPrefix("MTLS")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		l.Info("loading config: ok")
		dumpConfig()

		log.Setup(config.Config.GetBool(aconfig.SettingDebugLog))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		l.Fatal(err)
	}
}

func cmdServer(args *cli.Context) error {
	if err := validateConfig(config.Config); err != nil {
		l.Fatal(err)
	}

	backend := config.Config.GetString(
		aconfig.SettingMenderBackend,
	)

	proxy, err := api.NewProxy(backend)
	if err != nil {
		l.Fatal(err)
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
		l.Fatal(err)
	}

	app := app.NewApp(client, authProvider)
	r, err := api.NewRouter(app, proxy)
	if err != nil {
		l.Fatal(err)
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
		l.Fatal(err)
	}

	return s.Run()
}

func validateConfig(c config.Reader) error {
	l.Info("validating config")
	required := []string{
		aconfig.SettingMenderBackend,
		aconfig.SettingMenderUser,
		aconfig.SettingMenderPass,
	}

	for _, c := range required {
		if config.Config.GetString(c) == "" {
			return errors.New(fmt.Sprintf("validating config failed: need setting %s\n", c))
		}
	}

	l.Info("validating config: ok")
	return nil
}

func dumpConfig() {
	l.Info("config values:")
	l.Infof(" %s: %s",
		aconfig.SettingMenderBackend,
		config.Config.GetString(
			aconfig.SettingMenderBackend,
		))

	l.Infof(" %s: %s",
		aconfig.SettingMenderUser,
		config.Config.GetString(
			aconfig.SettingMenderUser,
		))

	pass := config.Config.GetString(
		aconfig.SettingMenderPass,
	)
	if pass != "" {
		l.Infof(" %s: %s", aconfig.SettingMenderPass, "not empty")
	} else {
		l.Infof(" %s: %s", aconfig.SettingMenderPass, "empty")
	}

	l.Infof(" %s: %s",
		aconfig.SettingServerCert,
		config.Config.GetString(
			aconfig.SettingServerCert,
		))

	l.Infof(" %s: %s",
		aconfig.SettingServerKey,
		config.Config.GetString(
			aconfig.SettingServerKey,
		))

	l.Infof(" %s: %s",
		aconfig.SettingServerKey,
		config.Config.GetString(
			aconfig.SettingTenantCAPem,
		))

	l.Infof(" %s: %s",
		aconfig.SettingListen,
		config.Config.GetString(
			aconfig.SettingListen,
		))

	l.Infof(" %s: %s",
		aconfig.SettingDebugLog,
		config.Config.GetString(
			aconfig.SettingDebugLog,
		))
}
