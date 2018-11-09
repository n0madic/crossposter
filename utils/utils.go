package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// IsRequestURL function
func IsRequestURL(rawurl string) bool {
	url, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return false
	}
	if len(url.Scheme) == 0 {
		return false
	}
	return true
}

// DownloadFile to disk
func DownloadFile(uri string, filePath string) error {
	writer, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()

	res, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", res.Status)
	}

	_, err = io.Copy(writer, res.Body)
	if err != nil {
		return err
	}

	return nil
}

// GetURLContentInBase64 get content from URL and return it in base64
func GetURLContentInBase64(uri string) (string, error) {
	res, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", res.Status)
	}

	buffer := new(bytes.Buffer)
	writer := base64.NewEncoder(base64.StdEncoding, buffer)
	_, err = io.Copy(writer, res.Body)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
