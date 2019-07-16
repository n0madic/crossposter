package main

import (
	"bytes"
	"html/template"
	"net/http"
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

	for _, consumer := range cfg.Consumers {
		if args.DontPost {
			consumer.Topics = []string{}
		}
		newConsumer, err := crossposter.Initializers[consumer.Type](consumer)
		if err != nil {
			log.Fatalln(err)
		}
		for _, topic := range consumer.Topics {
			crossposter.Events.SubscribeAsync(topic, newConsumer.Post, true)
		}
	}

	for _, producer := range cfg.Producers {
		newProducer, err := crossposter.Initializers[producer.Type](producer)
		if err != nil {
			log.Fatalln(err)
		}
		for _, source := range producer.Sources {
			crossposter.WaitGroup.Add(1)
			go newProducer.Get(source, lastUpdate)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("index").Parse(indexTpl)
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
