package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
	},
}

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
func DownloadFile(url string, filePath string) error {
	writer, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer writer.Close()

	res, err := http.Get(url)
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

// GetJSON from URL
func GetJSON(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Crossposter/1.0")

	r, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf(string(content))
	}

	return json.NewDecoder(r.Body).Decode(target)
}

// GetURLContentInBase64 get content from URL and return it in base64
func GetURLContentInBase64(url string) (string, error) {
	res, err := http.Get(url)
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

//NewDocumentToUTF8 return goquery Document in UTF-8 charset
func NewDocumentToUTF8(url, charset string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	utfBody, err := iconv.NewReader(res.Body, charset, "utf-8")
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromReader(utfBody)
}

// TruncateText is truncate strings to a fixed size
func TruncateText(text string, limit int) string {
	runeText := []rune(text)
	if len(runeText) <= limit {
		return text
	}
	return string(runeText[:limit-1]) + "…"
}

// StringInSlice check if string exists in the slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
