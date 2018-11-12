package rss

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/n0madic/crossposter"
)

// RSS entity
type RSS struct {
	feed      *feeds.Feed
	sourceURL string
}

func init() {
	crossposter.AddEntity("rss", New)
}

// New return RSS entity
func New(name string, entity crossposter.Entity) (crossposter.EntityInterface, error) {
	rss := &RSS{
		feed:      &feeds.Feed{Title: name},
		sourceURL: entity.URL,
	}
	http.HandleFunc("/rss/"+name, rss.Handler)
	return rss, nil
}

// Get items from RSS
func (rss *RSS) Get(name string) ([]crossposter.Post, error) {
	fp := gofeed.NewParser()
	sourceFeed, err := fp.ParseURL(rss.sourceURL)
	if err != nil {
		return nil, err
	}

	var posts []crossposter.Post
	mediaURLs := []string{}

	for _, item := range sourceFeed.Items {
		if item.Image != nil && item.Image.URL != "" {
			mediaURLs = append(mediaURLs, item.Image.URL)
		}
		for _, enclosure := range item.Enclosures {
			if strings.HasPrefix(enclosure.Type, "image/") {
				mediaURLs = append(mediaURLs, enclosure.URL)
			}
		}
		posts = append(posts, crossposter.Post{
			Date:        *item.PublishedParsed,
			URL:         item.Link,
			Author:      item.Author.Name,
			Title:       item.Title,
			Text:        item.Description,
			Attachments: mediaURLs,
			More:        false,
		})
	}

	return posts, nil
}

// Post add item to RSS feed
func (rss *RSS) Post(name string, post *crossposter.Post) (string, error) {
	description := post.Text
	for _, attach := range post.Attachments {
		description += fmt.Sprintf(`\n<br><img src="%s" />`, attach)
	}

	rss.feed.Add(&feeds.Item{
		Title:       post.Title,
		Link:        &feeds.Link{Href: post.URL},
		Description: description,
		Author:      &feeds.Author{Name: post.Author},
		Created:     post.Date,
	})
	return "", nil
}

// Handler return RSS XML
func (rss *RSS) Handler(w http.ResponseWriter, r *http.Request) {
	if len(rss.feed.Items) > 0 {
		xml, _ := rss.feed.ToRss()
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(xml))
	} else {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("No new RSS"))
	}
}
