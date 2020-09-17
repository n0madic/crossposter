package pikabu

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	lru "github.com/hashicorp/golang-lru"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
	log "github.com/sirupsen/logrus"
)

// Pikabu entity
type Pikabu struct {
	entity    *crossposter.Entity
	published *lru.Cache
}

var mutex sync.Mutex

func init() {
	crossposter.AddEntity("pikabu", New)
}

// New run Pikabu entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	cache, _ := lru.New(10000)
	return &Pikabu{entity: &entity, published: cache}, nil
}

// Get items from Pikabu
func (pikabu *Pikabu) Get(lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	for {
		for _, location := range pikabu.entity.Sources {
			pikabuLogger := log.WithFields(log.Fields{"location": location, "type": pikabu.entity.Type})
			pikabuLogger.Println("Check updates")

			mutex.Lock()
			doc, err := utils.NewDocumentToUTF8("https://pikabu.ru/"+location, "windows-1251")
			mutex.Unlock()
			if err != nil {
				pikabuLogger.Error(err)
			} else {
				posts := []crossposter.Post{}

				doc.Find(".story__main").Each(func(i int, sel *goquery.Selection) {
					sponsor := false
					sel.Find(".story__sponsor").Each(func(i int, c *goquery.Selection) {
						sponsor = true
					})
					timestamp, _ := time.Parse(time.RFC3339, sel.Find("div.user__info-item > time").First().AttrOr("datetime", ""))
					if !sponsor && !timestamp.IsZero() {
						var mediaURLs []string
						story := sel.Find(".story__content-inner").Each(func(i int, sel *goquery.Selection) {
							sel.Find("div.player").Each(func(i int, sel *goquery.Selection) {
								mediaURLs = append(mediaURLs, sel.AttrOr("data-source", ""))
							})
						})
						html, err := story.Html()
						if err != nil {
							html = story.Text()
						}
						post := crossposter.Post{
							Date:        timestamp,
							URL:         sel.Find(".story__title > a").First().AttrOr("href", ""),
							Author:      doc.Find(".user__nick").First().Text(),
							Title:       strings.TrimSpace(sel.Find(".story__title").First().Text()),
							Text:        strings.TrimSpace(html),
							Attachments: mediaURLs,
							More:        false,
						}
						posts = append(posts, post)
					}
				})

				sort.Slice(posts, func(i, j int) bool {
					return posts[i].Date.Before(posts[j].Date)
				})

				for _, post := range posts {
					if post.Date.After(lastUpdate) && !pikabu.published.Contains(post.URL) {
						lastUpdate = post.Date
						pikabu.published.Add(post.URL, nil)
						for _, topic := range pikabu.entity.Topics {
							crossposter.Events.Publish(topic, post)
							time.Sleep(time.Second * 5)
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(pikabu.entity.Wait) * time.Minute)
	}
}

// Post not implemented
func (pikabu *Pikabu) Post(post crossposter.Post) {}

// Handler not implemented
func (pikabu *Pikabu) Handler(w http.ResponseWriter, r *http.Request) {}
