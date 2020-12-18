package rss

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
	log "github.com/sirupsen/logrus"
)

const maxTitleLength = 50
const maxItemsInFeed = 10

// RSS entity
type RSS struct {
	entity *crossposter.Entity
	feed   *feeds.Feed
}

func init() {
	crossposter.AddEntity("rss", New)
}

// New run RSS entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	rss := &RSS{
		entity: &entity,
		feed: &feeds.Feed{
			Title:       entity.Options["title"],
			Description: entity.Description,
			Link:        &feeds.Link{Href: entity.Options["link"]},
		},
	}
	for _, destination := range entity.Destinations {
		http.HandleFunc("/rss/"+destination, rss.Handler)
	}
	return rss, nil
}

// Get items from RSS
func (rss *RSS) Get(lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()
	fp := gofeed.NewParser()

	for {
		for _, source := range rss.entity.Sources {
			rssLogger := log.WithFields(log.Fields{"source": source, "type": rss.entity.Type})
			rssLogger.Println("Check updates")
			sourceFeed, err := fp.ParseURL(source)
			if err != nil {
				rssLogger.Error(err)
			} else {
				if len(sourceFeed.Items) > 0 && sourceFeed.Items[0].PublishedParsed == nil {
					rssLogger.Error(fmt.Errorf("published datetime not found"))
					continue
				}

				sort.Slice(sourceFeed.Items, func(i, j int) bool {
					return sourceFeed.Items[i].PublishedParsed.Before(*sourceFeed.Items[j].PublishedParsed)
				})

				mediaURLs := []string{}
				for _, item := range sourceFeed.Items {
					if item.PublishedParsed.After(lastUpdate) {
						lastUpdate = *item.PublishedParsed
						if item.Image != nil && item.Image.URL != "" {
							mediaURLs = append(mediaURLs, item.Image.URL)
						}
						for _, enclosure := range item.Enclosures {
							if strings.HasPrefix(enclosure.Type, "image/") {
								mediaURLs = append(mediaURLs, enclosure.URL)
							}
						}
						author := ""
						if item.Author != nil {
							author = item.Author.Name
						}
						post := crossposter.Post{
							Date:        *item.PublishedParsed,
							URL:         item.Link,
							Author:      author,
							Title:       item.Title,
							Text:        item.Description,
							Attachments: mediaURLs,
							More:        false,
						}
						for _, topic := range rss.entity.Topics {
							crossposter.Events.Publish(topic, post)
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(rss.entity.Wait) * time.Minute)
	}
}

// Post add item to RSS feed
func (rss *RSS) Post(post crossposter.Post) {
	title := post.Title
	if title == "" {
		title = utils.TruncateText(post.Text, maxTitleLength)
	}

	description := post.Text
	for _, attach := range post.Attachments {
		description += fmt.Sprintf(`<br><img src="%s" />`, attach)
	}
	if post.More || title == "" {
		description += " " + post.URL
	}

	if len(rss.feed.Items) == maxItemsInFeed {
		rss.feed.Items = rss.feed.Items[1:]
	}

	rss.feed.Add(&feeds.Item{
		Title:       title,
		Link:        &feeds.Link{Href: post.URL},
		Description: strings.TrimSpace(description),
		Author:      &feeds.Author{Name: post.Author},
		Created:     post.Date,
	})
}

// Handler return RSS XML
func (rss *RSS) Handler(w http.ResponseWriter, r *http.Request) {
	if len(rss.feed.Items) > 0 {
		xml, err := rss.feed.ToRss()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(xml))
		}
	} else {
		w.Write([]byte("No new RSS"))
	}
}
