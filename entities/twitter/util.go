package twitter

import (
	"strings"
	"unicode/utf8"

	"github.com/n0madic/crossposter/utils"
)

// TwitterizeText prepares text for twitter status
func TwitterizeText(input string) string {
	if utf8.RuneCountInString(input) <= maxTweetLength {
		return input
	}
	truncatedText := ""
	currentTweetLength := 0
	maxAvailableLength := maxTweetLength - shortURLLength - 2
	words := strings.Fields(input)
	for _, word := range words {
		if utils.IsRequestURL(word) {
			currentTweetLength += shortURLLength
		} else {
			currentTweetLength += utf8.RuneCountInString(word)
		}
		if currentTweetLength < maxAvailableLength {
			truncatedText += word + " "
			currentTweetLength++
		} else {
			truncatedText += "…"
			break
		}
	}
	return strings.TrimSpace(truncatedText)
}
