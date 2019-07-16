package telegram

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/utils"
	log "github.com/sirupsen/logrus"
)

// Telegram entity
type Telegram struct {
	entity *crossposter.Entity
	client *tgbotapi.BotAPI
}

func init() {
	crossposter.AddEntity("telegram", New)
}

// New return Telegram entity
func New(entity crossposter.Entity) (crossposter.EntityInterface, error) {
	client, err := tgbotapi.NewBotAPI(entity.Options["token"])
	if err != nil {
		return nil, fmt.Errorf("failed to login: %v", err)
	}
	return &Telegram{&entity, client}, nil
}

// Get message from Telegram channel
func (tg *Telegram) Get(name string, lastUpdate time.Time) {
	defer crossposter.WaitGroup.Done()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tg.client.GetUpdatesChan(u)
	if err != nil {
		log.Panicln(err)
	}

	log.Printf("Check updates for [%s] %s", tg.entity.Type, name)
	for update := range updates {
		if update.ChannelPost == nil {
			continue
		}

		if utils.StringInSlice(update.ChannelPost.Chat.UserName, tg.entity.Sources) ||
			utils.StringInSlice(strconv.FormatInt(update.ChannelPost.Chat.ID, 10), tg.entity.Sources) {
			mediaURLs := []string{}
			if update.ChannelPost.Photo != nil {
				pixels := 0
				idx := -1
				for i, p := range *update.ChannelPost.Photo { // find largest image in set
					pix := p.Height * p.Width
					if pix > pixels {
						pixels = pix
						idx = i
					}
				}
				if idx != -1 {
					url, err := tg.client.GetFileDirectURL((*update.ChannelPost.Photo)[idx].FileID)
					if err != nil {
						log.Panicln(err)
					}
					mediaURLs = append(mediaURLs, url)
				}
			}

			url := ""
			if update.ChannelPost.Chat.UserName != "" {
				url = fmt.Sprintf("https://t.me/%s/%v", update.ChannelPost.Chat.UserName, update.ChannelPost.MessageID)
			}

			username := ""
			if update.ChannelPost.From != nil {
				username = update.ChannelPost.From.UserName
			}

			post := crossposter.Post{
				Date:        time.Unix(int64(update.ChannelPost.Date), 0),
				URL:         url,
				Title:       update.ChannelPost.Caption,
				Author:      username,
				Text:        update.ChannelPost.Text,
				Attachments: mediaURLs,
			}
			for _, topic := range tg.entity.Topics {
				crossposter.Events.Publish(topic, post)
			}
		}
	}
}

// Post message to Telegram chat
func (tg *Telegram) Post(post crossposter.Post) {
	for _, destination := range tg.entity.Destinations {
		chatID, errID := strconv.ParseInt(destination, 10, 64)

		if post.Text != "" {
			var msg tgbotapi.MessageConfig
			text := post.Text + "\n" + post.URL
			if errID == nil {
				msg = tgbotapi.NewMessage(chatID, text)
			} else {
				if !strings.HasPrefix(destination, "@") {
					destination = "@" + destination
				}
				msg = tgbotapi.NewMessageToChannel(destination, text)
			}
			pmsg, err := tg.client.Send(msg)
			if err != nil {
				log.Error(err)
			} else {
				log.Printf("Posted https://t.me/%s/%v", pmsg.Chat.Title, pmsg.MessageID)
			}
		}

		if errID != nil && len(post.Attachments) > 0 {
			log.Error("Need ChatID for attachments")
		} else {
			for _, attach := range post.Attachments {
				res, err := http.Get(attach)
				if err != nil {
					log.Error(err)
					continue
				}
				defer res.Body.Close()

				if res.StatusCode != http.StatusOK {
					log.Errorf("bad status: %s for %s", res.Status, attach)
					continue
				}

				reader := tgbotapi.FileReader{Name: path.Base(attach), Reader: res.Body, Size: -1}
				msg := tgbotapi.NewPhotoUpload(chatID, reader)
				msg.Caption = post.Title

				pmsg, err := tg.client.Send(msg)
				if err != nil {
					log.Error(err)
				} else {
					log.Printf("Posted https://t.me/%s/%v", pmsg.Chat.Title, pmsg.MessageID)
				}
			}
		}
	}
}

// Handler not implemented
func (tg *Telegram) Handler(w http.ResponseWriter, r *http.Request) {}
