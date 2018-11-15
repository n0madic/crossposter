package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/config"
	_ "github.com/n0madic/crossposter/entities"
	log "github.com/sirupsen/logrus"
)

const timeLayout = "2006-01-02T15:04:05"

var (
	args struct {
		Bind     string `arg:"-b,env" help:"Bind address"`
		Config   string `arg:"env" help:"Config file"`
		DontPost bool   `arg:"-d,env" help:"Do not produce posts"`
		Last     string `arg:"-l,env" help:"Initial date for update"`
	}
	lastUpdate time.Time
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})
}

func main() {
	args.Bind = ":8000"
	args.Config = "config.yaml"
	args.Last = time.Now().Format(timeLayout)
	arg.MustParse(&args)

	cfg, err := config.New(args.Config)
	if err != nil {
		log.Fatalln(err)
	}

	lastUpdate, err = time.Parse(timeLayout, args.Last)
	if err != nil {
		log.Fatalf("Can't parse last update time: %s", err)
	}

	for _, entity := range cfg.Entities {
		if args.DontPost {
			entity.Topics = []string{}
		}
		newEntity, err := crossposter.Initializers[entity.Type](entity)
		if err != nil {
			log.Fatalln(err)
		}
		switch entity.Role {
		case "producer":
			for _, source := range entity.Sources {
				crossposter.WaitGroup.Add(1)
				go newEntity.Get(source, lastUpdate)
			}
		case "consumer":
			for _, topic := range entity.Topics {
				crossposter.Events.SubscribeAsync(topic, newEntity.Post, true)
			}
		default:
			log.Fatalf("%s role is not supported", entity.Role)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("index").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(indexTpl)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			var tpl bytes.Buffer
			err = t.Execute(&tpl, cfg)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			} else {
				w.Write(tpl.Bytes())
			}
		}
	})

	go func() {
		log.Fatal(http.ListenAndServe(args.Bind, nil))
	}()

	crossposter.WaitGroup.Wait()
}
