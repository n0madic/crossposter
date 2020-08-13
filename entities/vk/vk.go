package vk

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	vkapi "github.com/himidori/golang-vk-api"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
	log "github.com/sirupsen/logrus"
)

// Vk entity
type Vk struct {
	entity *crossposter.Entity
	client *vkapi.VKClient
	name   string
}

var userMap sync.Map

func init() {
	crossposter.AddEntity("vk", New)
}

// New return Vk entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	var client *vkapi.VKClient
	var err error
	if token, ok := entity.Options["token"]; ok {
		client, err = vkapi.NewVKClientWithToken(token, &vkapi.TokenOptions{
			ValidateOnStart: true,
			ServiceToken:    true,
		}, false)
	} else {
		client, err = vkapi.NewVKClient(vkapi.DeviceAndroid, entity.Options["user"], entity.Options["password"], false)
	}
	if err != nil {
		return nil, err
	}
	return &Vk{&entity, client, entity.Options["name"]}, nil
}

// Get posts from Vk wall
func (vk *Vk) Get(domain string, lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	for {
		log.Printf("Check updates for [%s] %s", vk.entity.Type, domain)
		Items, err := vk.client.WallGet(domain, 10, nil)
		if err != nil {
			log.Error(err)
			return
		}

		sort.Slice(Items.Posts, func(i, j int) bool {
			itime := time.Unix(Items.Posts[i].Date, 0)
			jtime := time.Unix(Items.Posts[j].Date, 0)
			return itime.Before(jtime)
		})

		for _, item := range Items.Posts {
			if item.MarkedAsAd == 0 {
				timestamp := time.Unix(item.Date, 0)
				if timestamp.After(lastUpdate) {
					lastUpdate = timestamp
					if item.CopyHistory != nil {
						item = item.CopyHistory[0]
					}
					var photos []string
					var needMore bool
					if item.Attachments != nil {
						if len(item.Attachments) > 1 {
							needMore = true
						}
						for _, attach := range item.Attachments {
							switch attach.Type {
							case "photo":
								photos = append(photos, getMaxSizePhoto(*attach.Photo))
							case "video":
								photos = append(photos, getMaxPreview(*attach.Video))
								needMore = true
							case "doc":
								if attach.Document.Type == 3 { // GIF
									photos = append(photos, attach.Document.URL)
									break
								}
							default:
								needMore = true
							}
						}
					}
					author, err := getNameFromID(vk.client, item.FromID)
					if err != nil {
						log.Error(err)
						return
					}
					post := crossposter.Post{
						Date:        timestamp,
						URL:         fmt.Sprintf("https://vk.com/wall%v_%v", item.FromID, item.ID),
						Author:      author,
						Text:        item.Text,
						Attachments: photos,
						More:        needMore,
					}
					for _, topic := range vk.entity.Topics {
						crossposter.Events.Publish(topic, post)
					}
				}
			}
		}
		time.Sleep(time.Duration(crossposter.WaitTime) * time.Minute)
	}
}

// Post to Vk
func (vk *Vk) Post(post crossposter.Post) {
	var mediaIDs []string

	screenName, err := vk.client.ResolveScreenName(vk.name)
	if err != nil {
		log.Error(err)
		return
	}
	if screenName.ObjectID == 0 {
		log.Errorf("public %s not found", vk.name)
	}

	for _, attach := range post.Attachments {
		filePath := path.Join(os.TempDir(), path.Base(attach))
		err := utils.DownloadFile(attach, filePath)
		if err != nil {
			log.Error(err)
			return
		}

		media, err := vk.client.UploadGroupWallPhotos(screenName.ObjectID, []string{filePath})
		if err != nil {
			log.Error(err)
			return
		}

		err = os.Remove(filePath)
		if err != nil {
			log.Error(err)
			return
		}
		mediaIDs = append(mediaIDs, vk.client.GetPhotosString(media))
	}

	message := post.Text
	if post.More {
		message += "\n" + post.URL
	}
	params := url.Values{}
	if len(mediaIDs) > 0 {
		params.Set("attachments", strings.Join(mediaIDs, ","))
	}
	postID, err := vk.client.WallPost(screenName.ObjectID, message, params)
	if err != nil {
		log.Error(err)
	} else {
		log.Printf("Posted in VK https://vk.com/wall-%v_%v", screenName.ObjectID, postID)
	}
}

// Handler not implemented
func (vk *Vk) Handler(w http.ResponseWriter, r *http.Request) {}

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

func getNameFromID(client *vkapi.VKClient, id int) (string, error) {
	name, ok := userMap.Load(id)
	if !ok {
		user, err := client.UsersGet([]int{int(^uint32(id))})
		if err != nil {
			return "", err
		}
		newname := user[0].FirstName + " " + user[0].LastName
		if user[0].Nickname != "" {
			newname += " aka " + user[0].Nickname
		}
		userMap.Store(id, strings.TrimSpace(newname))
		return newname, nil
	}
	return name.(string), nil
}
