package telegram

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	whitespaces = regexp.MustCompile(`(?m)^[ \t]+`)
	emptylines  = regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
)

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
