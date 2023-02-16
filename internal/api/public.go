package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func GetPublicIP() (string, error) {
	res, err := http.Get("https://checkip.amazonaws.com/")
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	return string(bytes.TrimSpace(body)), err
}
