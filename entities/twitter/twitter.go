package twitter

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
	log "github.com/sirupsen/logrus"
)

const shortURLLength = 23
const maxTweetLength = 140
const maxPhotoLimit = 4

// Twitter entity
type Twitter struct {
	entity *crossposter.Entity
	client *anaconda.TwitterApi
}

func init() {
	crossposter.AddEntity("twitter", New)
}

// New return Twitter entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	client := anaconda.NewTwitterApiWithCredentials(
		entity.Options["token"],
		entity.Options["token_secret"],
		entity.Options["key"],
		entity.Options["key_secret"],
	)
	if client == nil {
		return nil, fmt.Errorf("can't create new TwitterAPI")
	}
	return &Twitter{&entity, client}, nil
}

// Get user's timeline from Twitter
func (tw *Twitter) Get(screenName string, lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	twLogger := log.WithFields(log.Fields{"name": screenName, "type": tw.entity.Type})

	for {
		twLogger.Println("Check updates")
		v := url.Values{}
		v.Set("count", "10")
		v.Set("screen_name", screenName)

		tweets, err := tw.client.GetUserTimeline(v)
		if err != nil {
			twLogger.Error(err)
		} else {
			sort.Slice(tweets, func(i, j int) bool {
				itime, _ := tweets[i].CreatedAtTime()
				jtime, _ := tweets[j].CreatedAtTime()
				return itime.Before(jtime)
			})

			for _, tweet := range tweets {
				timestamp, _ := tweet.CreatedAtTime()
				if timestamp.After(lastUpdate) {
					lastUpdate = timestamp
					mediaURLs := []string{}
					for _, media := range tweet.Entities.Media {
						if media.Type == "photo" || media.Type == "animated_gif" {
							mediaURLs = append(mediaURLs, media.Media_url_https)
						}
					}
					post := crossposter.Post{
						Date:        timestamp,
						URL:         fmt.Sprintf("https://twitter.com/%s/status/%s", screenName, tweet.IdStr),
						Author:      tweet.User.ScreenName,
						Text:        tweet.FullText,
						Attachments: mediaURLs,
						More:        false,
					}
					for _, topic := range tw.entity.Topics {
						crossposter.Events.Publish(topic, post)
					}
				}
			}
		}
		time.Sleep(time.Duration(crossposter.WaitTime) * time.Minute)
	}
}

// Post status to Twitter
func (tw *Twitter) Post(post crossposter.Post) {
	var mediaIDs []string
	v := url.Values{}

	user, err := tw.client.GetSelf(v)
	twLogger := log.WithFields(log.Fields{"name": user.ScreenName, "type": tw.entity.Type})

	status := TwitterizeText(post.Text)
	if strings.HasSuffix(status, "…") || post.More {
		status += " " + post.URL
	}

	for index, attach := range post.Attachments {
		if b64, err := utils.GetURLContentInBase64(attach); err == nil {
			media, err := tw.client.UploadMedia(b64)
			if err != nil {
				twLogger.Error(err)
				return
			}
			mediaIDs = append(mediaIDs, media.MediaIDString)
		} else {
			twLogger.Error(err)
			return
		}
		if index == maxPhotoLimit-1 {
			break
		}
	}

	v.Set("media_ids", strings.Join(mediaIDs[:], ","))
	result, err := tw.client.PostTweet(strings.TrimSpace(status), v)
	if err != nil {
		twLogger.Error(err)
	} else {
		twLogger.Printf("Posted tweet https://twitter.com/%s/status/%s", result.User.ScreenName, result.IdStr)
	}
}

// Handler not implemented
func (tw *Twitter) Handler(w http.ResponseWriter, r *http.Request) {}
