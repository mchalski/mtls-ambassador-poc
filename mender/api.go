package mender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	LoginUrl   = "/api/management/v1/useradm/auth/login"
	PreauthUrl = "/api/management/v2/devauth/devices"

	ErrPreauthConflict = "preauth conflict:"
)

type Client struct {
	c *http.Client
}

func NewClient() *Client {
	return &Client{
		c: &http.Client{},
	}
}

func (client *Client) Login(user, pwd, host string) (string, error) {
	url := "https://" + host + LoginUrl
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
	default:
		return "", errors.New(fmt.Sprintf("unexpected response from login: HTTP %d\n%s", resp.StatusCode, string(body)))
	}
}

type PreauthReq struct {
	IdData map[string]interface{} `json:"identity_data"`
	PubKey string                 `json:"pubkey"`
}

func ErrIsPreauthConflict(e error) bool {
	return strings.HasPrefix(e.Error(), ErrPreauthConflict)
}

func (client *Client) Preauth(idData, pubKey, host, userToken string) error {
	url := "https://" + host + PreauthUrl

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
		return errors.New(ErrPreauthConflict + string(body))
	default:
		return errors.New(fmt.Sprintf("unexpected response from preauth: HTTP %d\n%s", resp.StatusCode, string(body)))
	}
}

type AuthReq struct {
	IdData      string `json:"id_data"`
	TenantToken string `json:"tenant_token"`
	PubKey      string `json:"pubkey"`
}
