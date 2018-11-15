package vk

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	vkapi "github.com/himidori/golang-vk-api"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
)

// Vk entity
type Vk struct {
	client *vkapi.VKClient
}

var userMap sync.Map

func init() {
	crossposter.AddEntity("vk", New)
}

// New return Vk entity
func New(name string, entity crossposter.Entity) (crossposter.EntityInterface, error) {
	var client *vkapi.VKClient
	var err error
	if token, ok := entity.Options["token"]; ok {
		client, err = vkapi.NewVKClientWithToken(token, &vkapi.TokenOptions{
			ValidateOnStart: true,
			ServiceToken:    true,
		})
	} else {
		client, err = vkapi.NewVKClient(vkapi.DeviceAndroid, entity.Options["user"], entity.Options["password"])
	}
	if err != nil {
		return nil, err
	}
	return &Vk{client}, nil
}

// Get posts from Vk wall
func (vk *Vk) Get(domain string) ([]crossposter.Post, error) {
	var posts []crossposter.Post
	Items, err := vk.client.WallGet(domain, 10, nil)
	if err != nil {
		return nil, err
	}
	for _, item := range Items.Posts {
		if item.MarkedAsAd == 0 {
			var photos []string
			var needMore bool
			if item.Attachments != nil {
				if len(item.Attachments) > 1 {
					needMore = true
				}
				for _, attach := range item.Attachments {
					switch attach.Type {
					case "photo":
						photos = append(photos, GetMaxSizePhoto(*attach.Photo))
					case "video":
						photos = append(photos, GetMaxPreview(*attach.Video))
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
				return nil, err
			}
			posts = append(posts, crossposter.Post{
				Date:        time.Unix(item.Date, 0),
				URL:         fmt.Sprintf("https://vk.com/wall%v_%v", item.FromID, item.ID),
				Author:      author,
				Text:        item.Text,
				Attachments: photos,
				More:        needMore,
			})
		}
	}

	return posts, nil
}

// Post to Vk
func (vk *Vk) Post(name string, post *crossposter.Post) (string, error) {
	var mediaIDs []string

	screenName, err := vk.client.ResolveScreenName(name)
	if err != nil {
		return "", err
	}
	if screenName.ObjectID == 0 {
		return "", fmt.Errorf("public %s not found", name)
	}

	for _, attach := range post.Attachments {
		filePath := path.Join(os.TempDir(), path.Base(attach))
		err := utils.DownloadFile(attach, filePath)
		if err != nil {
			return "", err
		}

		media, err := vk.client.UploadGroupWallPhotos(screenName.ObjectID, []string{filePath})
		if err != nil {
			return "", err
		}

		err = os.Remove(filePath)
		if err != nil {
			return "", err
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
		return "", err
	}

	return fmt.Sprintf("Posted in VK https://vk.com/wall-%v_%v", screenName.ObjectID, postID), nil
}

// Handler not implemented
func (vk *Vk) Handler(w http.ResponseWriter, r *http.Request) {}

// GetMaxSizePhoto from attachment
func GetMaxSizePhoto(p vkapi.PhotoAttachment) string {
	if p.Photo2560 != "" {
		return p.Photo2560
	}
	if p.Photo1280 != "" {
		return p.Photo1280
	}
	if p.Photo807 != "" {
		return p.Photo807
	}
	if p.Photo604 != "" {
		return p.Photo604
	}
	if p.Photo130 != "" {
		return p.Photo130
	}
	if p.Photo75 != "" {
		return p.Photo75
	}

	return ""
}

// GetMaxPreview from video attachment
func GetMaxPreview(v vkapi.VideoAttachment) string {
	if v.Photo800 != "" {
		return v.Photo800
	}
	if v.FirstFrame800 != "" {
		return v.FirstFrame800
	}
	if v.Photo320 != "" {
		return v.Photo320
	}
	if v.FirstFrame320 != "" {
		return v.FirstFrame320
	}
	if v.FirstFrame160 != "" {
		return v.FirstFrame160
	}
	if v.Photo130 != "" {
		return v.Photo130
	}
	if v.FirstFrame130 != "" {
		return v.FirstFrame130
	}

	return ""
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
