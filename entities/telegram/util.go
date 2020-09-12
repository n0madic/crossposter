package telegram

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/n0madic/crossposter"
	log "github.com/sirupsen/logrus"
)

var (
	whitespaces = regexp.MustCompile(`(?m)^[ \t]+`)
	emptylines  = regexp.MustCompile(`(?m)^[ \t]+|^\s*$[\r\n]*|[\r\n]+\s+\z`)
)

func getText(post *crossposter.Post) string {
	switch {
	case post.Title != "" && post.URL != "":
		return fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n%s", post.URL, post.Title, post.Text)
	case post.Title != "" && post.URL == "":
		return fmt.Sprintf("<b>%s</b>\n%s", post.Title, post.Text)
	case post.Title == "" && post.URL != "":
		return fmt.Sprintf("<b><a href=\"%s\">%s</a></b>\n%s", post.URL, post.URL, post.Text)
	}
	return post.Text
}

func extractImages(post *crossposter.Post) {
	base, err := url.Parse(post.URL)
	if err != nil {
		log.Error(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(post.Text))
	if err != nil {
		log.Error(err)
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
}

func sanitize(html string) string {
	replacer := strings.NewReplacer(
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
	)
	html = replacer.Replace(html)
	html = strings.ReplaceAll(html, "</p>", "</p>\n")

	p := bluemonday.NewPolicy()
	p.AllowStandardURLs()
	p.AllowAttrs("href").OnElements("a")
	p.AllowElements(
		"b", "strong",
		"i", "em",
		"u", "ins",
		"s", "strike", "del",
		"code", "pre",
	)

	html = p.Sanitize(html)
	html = whitespaces.ReplaceAllString(html, "")
	html = emptylines.ReplaceAllString(html, "")
	return strings.TrimSpace(html)
}
