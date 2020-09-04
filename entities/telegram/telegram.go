package telegram

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	md "github.com/JohannesKaufmann/html-to-markdown"
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

	tgLogger := log.WithFields(log.Fields{"channel": name, "type": tg.entity.Type})

	updates, err := tg.client.GetUpdatesChan(u)
	if err != nil {
		tgLogger.Panicln(err)
	}

	tgLogger.Println("Check updates")
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

		tgLogger := log.WithFields(log.Fields{"chat": destination, "type": tg.entity.Type})

		converter := md.NewConverter("", true, nil)
		text, err := converter.ConvertString(post.Text)
		if err != nil {
			text = post.Text
		} else {
			text = strings.ReplaceAll(text, ` \- `, ` - `)
		}
		switch {
		case post.Title != "" && post.URL != "":
			text = fmt.Sprintf("[%s](%s)\n%s", post.Title, post.URL, text)
		case post.Title != "" && post.URL == "":
			text = fmt.Sprintf("*%s*\n%s", post.Title, text)
		case post.Title == "" && post.URL != "":
			text = fmt.Sprintf("%s\n%s", post.URL, text)
		}

		if (text != "" && len(post.Attachments) == 0) || utf8.RuneCountInString(text) > 1024 {
			if utf8.RuneCountInString(text) > 4096 {
				text = utils.TruncateText(text, 4096)
			}

			var msg tgbotapi.MessageConfig
			if errID == nil {
				msg = tgbotapi.NewMessage(chatID, text)
			} else {
				if !strings.HasPrefix(destination, "@") {
					destination = "@" + destination
				}
				msg = tgbotapi.NewMessageToChannel(destination, text)
			}
			msg.ParseMode = "Markdown"

			pmsg, err := tg.client.Send(msg)
			if err != nil {
				tgLogger.Error(err)
			} else {
				tgLogger.Printf("Posted https://t.me/%s/%v", pmsg.Chat.Title, pmsg.MessageID)
			}
		}

		if len(post.Attachments) > 0 {
			if errID != nil {
				tgLogger.Error("Need ChatID for post attachments")
			} else {
				var files []interface{}
				for i := 0; i < 10 && i < len(post.Attachments); i++ {
					files = append(files, tgbotapi.NewInputMediaPhoto(post.Attachments[i]))
				}
				if utf8.RuneCountInString(text) <= 1024 {
					files[0] = tgbotapi.InputMediaPhoto{
						Type:      "photo",
						Media:     post.Attachments[0],
						Caption:   text,
						ParseMode: "Markdown",
					}
				}
				msg := tgbotapi.NewMediaGroup(chatID, files)

				pmsg, err := tg.client.Send(msg)
				if err != nil {
					tgLogger.Error(err)
				} else {
					if pmsg.Chat != nil {
						tgLogger.Printf("Posted https://t.me/%s/%v", pmsg.Chat.Title, pmsg.MessageID)
					} else {
						tgLogger.Printf("Posted media group")
					}
				}
			}
		}
	}
}

// Handler not implemented
func (tg *Telegram) Handler(w http.ResponseWriter, r *http.Request) {}
