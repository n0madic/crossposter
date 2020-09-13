package vk

import (
	"regexp"

	vkapi "github.com/himidori/golang-vk-api"
)

var reInternalURLs = regexp.MustCompile(`\[(.+?)\|(.+?)\]`)

// getMaxSizePhoto from attachment
func getMaxSizePhoto(p vkapi.PhotoAttachment) string {
	maxWidth := 0
	url := ""
	for _, photo := range p.Sizes {
		if photo.Width > maxWidth {
			maxWidth = photo.Width
			url = photo.Url
		}
	}
	return url
}

// getMaxPreview from video attachment
func getMaxPreview(v vkapi.VideoAttachment) string {
	maxWidth := 0
	url := ""
	for _, image := range v.Image {
		if image.Width > maxWidth {
			maxWidth = image.Width
			url = image.Url
		}
	}
	return url
}
