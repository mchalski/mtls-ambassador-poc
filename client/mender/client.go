// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package mender

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mendersoftware/go-lib-micro/log"
)

const (
	LoginUrl   = "/api/management/v1/useradm/auth/login"
	PreauthUrl = "/api/management/v2/devauth/devices"

	DefaultTimeoutSec = 5
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrPreauthConflict = errors.New("preauth conflict")
	l                  = log.NewEmpty()
)

type Client interface {
	Login(ctx context.Context, user, pwd string) (string, error)
	Preauth(ctx context.Context, idData, pubKey, userToken string) error
}

type client struct {
	c       *http.Client
	baseUrl string
}

func NewClient(baseUrl string) *client {
	l.Infof("created client with base url %s", baseUrl)
	return &client{
		c: &http.Client{
			Timeout: DefaultTimeoutSec * time.Second,
		},
		baseUrl: baseUrl,
	}
}

func (client *client) Login(ctx context.Context, user, pwd string) (string, error) {
	url := join(client.baseUrl, LoginUrl)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(user, pwd)
	resp, err := client.c.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return string(body), nil
	case http.StatusUnauthorized:
		return "", ErrUnauthorized
	default:
		return "", errors.New(fmt.Sprintf("unexpected response from login: HTTP %d\n%s", resp.StatusCode, body))
	}
}

func (client *client) Preauth(ctx context.Context, idData, pubKey, userToken string) error {
	url := join(client.baseUrl, PreauthUrl)

	var idAttrs map[string]interface{}

	err := json.Unmarshal([]byte(idData), &idAttrs)
	if err != nil {
		return err
	}

	preauthReq := PreauthReq{
		IdData: idAttrs,
		PubKey: pubKey,
	}

	preauthBody, err := json.Marshal(preauthReq)
	if err != nil {
		return err
	}

	rd := bytes.NewReader(preauthBody)

	req, err := http.NewRequest(http.MethodPost, url, rd)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusConflict:
		return ErrPreauthConflict
	default:
		return errors.New(fmt.Sprintf("unexpected response from preauth: HTTP %d\n%s", resp.StatusCode, body))
	}
}

func join(base, url string) string {
	if strings.HasPrefix(url, "/") {
		url = url[1:]
	}
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}
	return base + url
}
