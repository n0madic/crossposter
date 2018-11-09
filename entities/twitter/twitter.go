package twitter

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"

	"github.com/ChimeraCoder/anaconda"
)

const shortURLLength = 23
const maxTweetLength = 140
const maxPhotoLimit = 4

// Twitter entity
type Twitter struct {
	client *anaconda.TwitterApi
}

func init() {
	crossposter.AddEntity("twitter", New)
}

// New return Twitter entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	client := anaconda.NewTwitterApiWithCredentials(
		entity.Token,
		entity.TokenSecret,
		entity.Key,
		entity.KeySecret,
	)
	if client == nil {
		return nil, fmt.Errorf("can't create new TwitterAPI")
	}
	return &Twitter{client}, nil
}

// Get user's timeline from Twitter
func (tw *Twitter) Get(screenName string) ([]crossposter.Post, error) {
	v := url.Values{}

	v.Set("count", "10")
	v.Set("screen_name", screenName)

	tweets, err := tw.client.GetUserTimeline(v)
	if err != nil {
		return nil, err
	}
	var posts []crossposter.Post
	for _, tweet := range tweets {
		timestamp, _ := tweet.CreatedAtTime()
		mediaURLs := []string{}
		for _, media := range tweet.Entities.Media {
			if media.Type == "photo" || media.Type == "animated_gif" {
				mediaURLs = append(mediaURLs, media.Media_url_https)
			}
		}
		posts = append(posts, crossposter.Post{
			Date:        timestamp,
			URL:         fmt.Sprintf(" https://twitter.com/%s/status/%s", screenName, tweet.IdStr),
			Author:      tweet.User.ScreenName,
			Text:        tweet.FullText,
			Attachments: mediaURLs,
			More:        false,
		})
	}

	return posts, nil
}

// Post status to Twitter
func (tw *Twitter) Post(name string, post *crossposter.Post) (string, error) {
	var mediaIDs []string

	status := TwitterizeText(post.Text)
	if strings.HasSuffix(status, "…") || post.More {
		status += post.URL
	}

	for index, attach := range post.Attachments {
		if b64, err := utils.GetURLContentInBase64(attach); err != nil {
			media, err := tw.client.UploadMedia(b64)
			if err != nil {
				return "", err
			}
			mediaIDs = append(mediaIDs, media.MediaIDString)
		} else {
			return "", err
		}
		if index == maxPhotoLimit-1 {
			break
		}
	}

	v := url.Values{}
	v.Set("media_ids", strings.Join(mediaIDs[:], ","))
	result, err := tw.client.PostTweet(strings.TrimSpace(status), v)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Posted tweet https://twitter.com/%s/status/%s", result.User.ScreenName, result.IdStr), nil
}

// TwitterizeText prepares text for twitter status
func TwitterizeText(input string) string {
	if utf8.RuneCountInString(input) <= maxTweetLength {
		return input
	}
	truncatedText := ""
	currentTweetLength := 0
	maxAvailableLength := maxTweetLength - shortURLLength - 2
	reTags := regexp.MustCompile(`\[(club|id)\d+\|(.+)\]`)
	words := strings.Fields(reTags.ReplaceAllString(input, ""))
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
