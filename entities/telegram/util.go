package telegram

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/n0madic/crossposter"
)

var (
	whitespaces = regexp.MustCompile(`(?m)^[ \t]+`)
	emptylines  = regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
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

func sanitize(html string) string {
	replacer := strings.NewReplacer(
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
	)
	html = replacer.Replace(html)

	for _, tag := range []string{"</p>", "</blockquote>"} {
		html = strings.ReplaceAll(html, tag, tag+"\n")
	}

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
