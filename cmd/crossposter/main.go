package main

import (
	"bytes"
	"html/template"
	"net/http"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/n0madic/crossposter"
	"github.com/n0madic/crossposter/config"
	log "github.com/sirupsen/logrus"
)

const timeLayout = "2006-01-02T15:04:05"

var (
	args struct {
		Bind     string `arg:"-b,env" help:"Bind address" default:":8000"`
		Config   string `arg:"-c,env" help:"Config file" default:"config.yaml"`
		DontPost bool   `arg:"-d,env:DONT_POST" help:"Do not post"`
		Last     string `arg:"-i,env" help:"Initial date for update"`
		LogLevel string `arg:"-l,env:LOG_LEVEL" help:"Set log level" default:"info"`
	}
	lastUpdate time.Time
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})
	args.Last = time.Now().Format(timeLayout)
}

func main() {
	arg.MustParse(&args)

	ll, err := log.ParseLevel(args.LogLevel)
	if err != nil {
		log.Fatalf("Can't parse log level: %s", err)
	}
	log.SetLevel(ll)

	cfg, err := config.New(args.Config)
	if err != nil {
		log.Fatalln(err)
	}

	lastUpdate, err = time.Parse(timeLayout, args.Last)
	if err != nil {
		log.Fatalf("Can't parse last update time: %s", err)
	}

	err = cfg.SubscribeConsumers(args.DontPost)
	if err != nil {
		log.Fatalln(err)
	}

	err = cfg.RunProducers(lastUpdate)
	if err != nil {
		log.Fatalln(err)
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
