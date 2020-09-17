package crossposter

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Post data struct
type Post struct {
	Date        time.Time
	URL         string
	Author      string
	Title       string
	Text        string
	Attachments []string
	More        bool
}

// ExtractImages from HTML to attachments
func (post *Post) ExtractImages() error {
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
	}).Remove()
	doc.Find("a:empty").Remove()

	post.Text, err = doc.Html()
	return err
}

func (post *Post) FullText() string {
	switch {
	case post.Title != "" && post.URL != "":
		return fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n%s", post.URL, post.Title, post.Text)
	case post.Title != "" && post.URL == "":
		return fmt.Sprintf("<b>%s</b>\n%s", post.Title, post.Text)
	case post.Title == "" && post.URL != "":
		return fmt.Sprintf("<a href=\"%s\">%s</a>\n%s", post.URL, post.URL, post.Text)
	}
	return post.Text
}
