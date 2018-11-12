package rss

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
)

const maxTitleLength = 50

// RSS entity
type RSS struct {
	feed *feeds.Feed
}

func init() {
	crossposter.AddEntity("rss", New)
}

// New return RSS entity
func New(name string, entity crossposter.Entity) (crossposter.EntityInterface, error) {
	rss := &RSS{
		feed: &feeds.Feed{
			Title:       name,
			Description: entity.Description,
			Link:        &feeds.Link{Href: entity.URL},
		},
	}
	http.HandleFunc("/rss/"+name, rss.Handler)
	return rss, nil
}

// Get items from RSS
func (rss *RSS) Get(name string) ([]crossposter.Post, error) {
	fp := gofeed.NewParser()
	sourceFeed, err := fp.ParseURL(rss.feed.Link.Href)
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

	rss.feed.Add(&feeds.Item{
		Title:       title,
		Link:        &feeds.Link{Href: post.URL},
		Description: strings.TrimSpace(description),
		Author:      &feeds.Author{Name: post.Author},
		Created:     post.Date,
	})
	return "/rss/" + name, nil
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
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("No new RSS"))
	}
}
