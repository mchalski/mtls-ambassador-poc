package mender

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	LoginUrl = "/api/management/v1/useradm/auth/login"
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
