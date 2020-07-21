// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package config

import (
	"github.com/mendersoftware/go-lib-micro/config"
)

const (
	// SettingListen is the config key for the listen port
	SettingListen        = "listen"
	SettingListenDefault = ":8080"

	// SettingMenderBackend is the config key for the Mender base url (scheme + host:port)
	SettingMenderBackend        = "mender_backend"
	SettingMenderBackendDefault = ""

	// SettingMenderUser is the Ambassador's Mender login
	SettingMenderUser        = "mender_user"
	SettingMenderUserDefault = ""

	// SettingMenderPass is the Ambassador's Mender login
	SettingMenderPass        = "mender_pass"
	SettingMenderPassDefault = ""

	// SettingServerCert is the Ambassador's https server cert
	SettingServerCert        = "server_cert"
	SettingServerCertDefault = "/etc/mtls/certs/server/server.crt"

	// SettingServerKey is the Ambassador's https server key
	SettingServerKey        = "server_key"
	SettingServerKeyDefault = "/etc/mtls/certs/server/server.key"

	// SettingTenantCAPem is the tenant's CA cert used to verify client mTLS certs
	SettingTenantCAPem        = "tenant_ca_pem"
	SettingTenantCAPemDefault = "/etc/mtls/certs/tenant-ca/tenant.ca.pem"

	// SettingDebugLog is the config key for the turning on the debug log
	SettingDebugLog        = "debug_log"
	SettingDebugLogDefault = false
)

var (
	// Defaults are the default configuration settings
	Defaults = []config.Default{
		{Key: SettingListen, Value: SettingListenDefault},
		{Key: SettingMenderBackend, Value: SettingMenderBackendDefault},
		{Key: SettingMenderUser, Value: SettingMenderUserDefault},
		{Key: SettingMenderPass, Value: SettingMenderPassDefault},
		{Key: SettingServerCert, Value: SettingServerCertDefault},
		{Key: SettingServerKey, Value: SettingServerKeyDefault},
		{Key: SettingTenantCAPem, Value: SettingTenantCAPemDefault},
		{Key: SettingDebugLog, Value: SettingDebugLogDefault},
	}
)
