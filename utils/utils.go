package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	"github.com/n0madic/crossposter"
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

// ExtractImages from HTML to attachments
func ExtractImages(post *crossposter.Post) error {
	base, err := url.Parse(post.URL)
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(post.Text))
	if err != nil {
		return err
	}
	doc.Find("img").Each(func(i int, sel *goquery.Selection) {
		for _, attr := range []string{"src", "data-src"} {
			src, exist := sel.Attr(attr)
			if exist && src != "" {
				u, err := url.Parse(src)
				if err == nil {
					if !u.IsAbs() {
						src = base.ResolveReference(u).String()
					}
					post.Attachments = append(post.Attachments, src)
				}
			}
		}
	})
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
