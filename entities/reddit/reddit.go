package reddit

import (
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
	log "github.com/sirupsen/logrus"
)

// Reddit entity
type Reddit struct {
	entity    *crossposter.Entity
	published *lru.Cache
}

type (
	Submission struct {
		Author        string        `json:"author"`
		CreatedUTC    jsonTimestamp `json:"created_utc"`
		LinkFlairText string        `json:"link_flair_text"`
		Permalink     string        `json:"permalink"`
		Pinned        bool          `json:"pinned"`
		PostHint      string        `json:"post_hint"`
		SelftextHTML  string        `json:"selftext_html"`
		Stickied      bool          `json:"stickied"`
		Title         string        `json:"title"`
		URL           string        `json:"url"`
	}

	Subreddit struct {
		Data struct {
			Children []struct {
				Data Submission `json:"data,omitempty"`
			} `json:"children"`
		} `json:"data"`
	}
)

func init() {
	crossposter.AddEntity("reddit", New)
}

// New return reddit entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	cache, err := lru.New(10000)
	return &Reddit{entity: &entity, published: cache}, err
}

// Get reddit message
func (reddit *Reddit) Get(lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	for {
		for _, name := range reddit.entity.Sources {
			redLogger := log.WithFields(log.Fields{"sub": name, "type": reddit.entity.Type})
			redLogger.Info("Check subreddit updates")

			posts := []crossposter.Post{}
			var data Subreddit
			url := fmt.Sprintf("https://www.reddit.com/r/%s.json", name)
			err := utils.GetJSON(url, &data)
			if err != nil {
				redLogger.Error(err)
			} else {
				for _, sub := range data.Data.Children {
					if !sub.Data.Pinned && !sub.Data.Stickied && sub.Data.LinkFlairText != "MOD POST" {
						var mediaURLs []string
						url := sub.Data.URL
						switch sub.Data.PostHint {
						case "image":
							mediaURLs = append(mediaURLs, sub.Data.URL)
							url = "https://www.reddit.com" + sub.Data.Permalink
						}
						text := html.UnescapeString(sub.Data.SelftextHTML)
						text = strings.TrimPrefix(text, "<!-- SC_OFF -->")
						text = strings.TrimSuffix(text, "<!-- SC_ON -->")
						post := crossposter.Post{
							Date:        time.Time(sub.Data.CreatedUTC),
							URL:         url,
							Author:      sub.Data.Author,
							Title:       sub.Data.Title,
							Text:        strings.TrimSpace(text),
							Attachments: mediaURLs,
							More:        true,
						}
						posts = append(posts, post)
					}
				}

				sort.Slice(posts, func(i, j int) bool {
					return posts[i].Date.Before(posts[j].Date)
				})

				for _, post := range posts {
					if post.Date.After(lastUpdate) && !reddit.published.Contains(post.URL) {
						lastUpdate = post.Date
						reddit.published.Add(post.URL, nil)
						for _, topic := range reddit.entity.Topics {
							crossposter.Events.Publish(topic, post)
							time.Sleep(time.Second * 5)
						}
					}
				}
			}
		}
		time.Sleep(time.Duration(reddit.entity.Wait) * time.Minute)
	}
}

// Post reddit message
func (reddit *Reddit) Post(post crossposter.Post) {}

// Handler reddit message
func (reddit *Reddit) Handler(w http.ResponseWriter, r *http.Request) {}
