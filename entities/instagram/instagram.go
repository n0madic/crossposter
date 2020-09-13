package instagram

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/n0madic/crossposter"
	log "github.com/sirupsen/logrus"
	goinsta "gopkg.in/ahmdrz/goinsta.v2"
)

// Instagram entity
type Instagram struct {
	entity *crossposter.Entity
	client *goinsta.Instagram
}

func init() {
	crossposter.AddEntity("instagram", New)
}

// New return Instagram entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	client := goinsta.New(
		entity.Options["user"],
		entity.Options["password"],
	)
	if err := client.Login(); err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	return &Instagram{&entity, client}, nil
}

// Get user's feed from Instagram
func (inst *Instagram) Get(name string, lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	insLogger := log.WithFields(log.Fields{"name": name, "type": inst.entity.Type})

	for {
		insLogger.Println("Check updates")
		user, err := inst.client.Profiles.ByName(name)
		if err != nil {
			insLogger.Error(err)
		} else {
			media := user.Feed()
			media.Next()

			sort.Slice(media.Items, func(i, j int) bool {
				itime := time.Unix(int64(media.Items[i].TakenAt), 0)
				jtime := time.Unix(int64(media.Items[j].TakenAt), 0)
				return itime.Before(jtime)
			})

			for _, item := range media.Items {
				itime := time.Unix(int64(item.TakenAt), 0)
				if itime.After(lastUpdate) {
					lastUpdate = itime
					mediaURLs := []string{}
					if item.Images.GetBest() != "" {
						mediaURLs = append(mediaURLs, item.Images.GetBest())
					}
					for _, slide := range item.CarouselMedia {
						mediaURLs = append(mediaURLs, slide.Images.GetBest())
					}
					post := crossposter.Post{
						Date:        time.Unix(int64(item.TakenAt), 0),
						URL:         fmt.Sprintf("https://www.instagram.com/p/%s", item.Code),
						Author:      user.FullName,
						Text:        item.Caption.Text,
						Attachments: mediaURLs,
						More:        item.MediaToString() != "photo",
					}
					for _, topic := range inst.entity.Topics {
						crossposter.Events.Publish(topic, post)
					}

				}
			}
		}
		time.Sleep(time.Duration(inst.entity.Wait) * time.Minute)
	}
}

// Post media to Instagram
func (inst *Instagram) Post(post crossposter.Post) {
	insLogger := log.WithFields(log.Fields{"type": inst.entity.Type})

	for _, attach := range post.Attachments {
		res, err := http.Get(attach)
		if err != nil {
			insLogger.Error(err)
			return
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			insLogger.Errorf("bad status: %s", res.Status)
		}

		caption := post.Text
		if post.More {
			caption += "\n" + post.URL
		}

		item, err := inst.client.UploadPhoto(res.Body, caption, 82, 0)
		if err != nil {
			insLogger.Error(err)
		} else {
			insLogger.Printf("Posted https://www.instagram.com/p/%s", item.Code)
		}
	}
}

// Handler not implemented
func (inst *Instagram) Handler(w http.ResponseWriter, r *http.Request) {}
