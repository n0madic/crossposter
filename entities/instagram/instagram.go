package instagram

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/n0madic/crossposter"
	goinsta "gopkg.in/ahmdrz/goinsta.v2"
)

// Instagram entity
type Instagram struct {
	client *goinsta.Instagram
}

func init() {
	crossposter.AddEntity("instagram", New)
}

// New return Instagram entity
func New(name string, entity crossposter.Entity) (crossposter.EntityInterface, error) {
	client := goinsta.New(
		entity.Key,
		entity.KeySecret,
	)
	if err := client.Login(); err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	return &Instagram{client}, nil
}

// Get user's feed from Instagram
func (inst *Instagram) Get(name string) ([]crossposter.Post, error) {
	user, err := inst.client.Profiles.ByName(name)
	if err != nil {
		return nil, err
	}

	var posts []crossposter.Post

	media := user.Feed()
	media.Next()

	for _, item := range media.Items {
		posts = append(posts, crossposter.Post{
			Date:        time.Unix(int64(item.TakenAt), 0),
			URL:         fmt.Sprintf("https://www.instagram.com/p/%s", item.Code),
			Author:      user.FullName,
			Text:        item.Caption.Text,
			Attachments: []string{item.Images.GetBest()},
			More:        item.MediaToString() != "photo",
		})
	}

	return posts, nil
}

// Post media to Instagram
func (inst *Instagram) Post(name string, post *crossposter.Post) (string, error) {
	var mediaURLs []string

	for _, attach := range post.Attachments {
		res, err := http.Get(attach)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return "", fmt.Errorf("bad status: %s", res.Status)
		}

		caption := post.Text
		if post.More {
			caption += "\n" + post.URL
		}

		item, err := inst.client.UploadPhoto(res.Body, caption, 82, 0)
		if err != nil {
			return "", err
		}
		mediaURLs = append(mediaURLs, fmt.Sprintf("https://www.instagram.com/p/%s", item.Code))
	}

	return strings.Join(mediaURLs, " "), nil
}

// Handler not implemented
func (inst *Instagram) Handler(w http.ResponseWriter, r *http.Request) {}
