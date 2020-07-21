// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package mender

type AuthReq struct {
	IdData      string `json:"id_data"`
	TenantToken string `json:"tenant_token"`
	PubKey      string `json:"pubkey"`
}

type PreauthReq struct {
	IdData map[string]interface{} `json:"identity_data"`
	PubKey string                 `json:"pubkey"`
}
