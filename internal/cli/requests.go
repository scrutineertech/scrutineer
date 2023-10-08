package cli

import (
	"fmt"
	"io"
	"net/http"
)

func (c Cli) userAgent() string {
	return fmt.Sprintf("scrutineer/%s", c.Version)
}

func (c Cli) getRequest(urlPath string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, c.serverUrl.JoinPath(urlPath).String(), nil)
	if err != nil {
		return nil, err
	}

	authCache, err := c.getAuthCache()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authCache.Token)

	req.Header.Add("User-Agent", c.userAgent())
	return req, nil
}

func (c Cli) postRequest(path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, c.serverUrl.JoinPath(path).String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	authCache, err := c.getAuthCache()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authCache.Token)

	req.Header.Add("User-Agent", c.userAgent())
	return req, nil
}

func (c Cli) deleteRequest(path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodDelete, c.serverUrl.JoinPath(path).String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	authCache, err := c.getAuthCache()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authCache.Token)

	req.Header.Add("User-Agent", c.userAgent())
	return req, nil
}

func (c Cli) postUnauthRequest(path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, c.serverUrl.JoinPath(path).String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("User-Agent", c.userAgent())
	return req, nil
}
