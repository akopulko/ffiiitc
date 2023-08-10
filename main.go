package main

import (
	"ffiiitc/internal/classifier"
	"ffiiitc/internal/config"
	"ffiiitc/internal/firefly"
	"ffiiitc/internal/handlers"
	"ffiiitc/internal/router"
	"log"

	"github.com/go-pkgz/lgr"
)

const (
	ffAppTimeout = 10
)

func main() {

	l := lgr.New(lgr.Msec, lgr.Debug, lgr.CallerFile, lgr.CallerFunc)

	// get the config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	// make firefly http client for rest api
	fc := firefly.NewFireFlyHttpClient(cfg.FFApp, cfg.APIKey, ffAppTimeout, l)

	// make classifier
	// on first run, classifier will take all your
	// transactions and learn their categories
	// subsequent start classifier will load trained model from file
	cls := classifier.NewTrnClassifier(fc, l)

	log.Printf("Learned classes:\n %v", cls.Classifier.Classes)

	h := handlers.NewWebHookHandler(cls, fc, l)

	r := router.NewRouter()
	r.AddRoute("/classify", h.HandleNewTransactionWebHook)
	r.AddRoute("/learn", h.HandleUpdateTransactionWebHook)
	err = r.Run(8080)
	if err != nil {
		panic(err)
	}
	// http.HandleFunc("/", HandleNewTransactionWebHook(cls, fc))
	// http.HandleFunc("/learn", HandleUpdateTransactionWebHook(cls))
	//log.Fatal(http.ListenAndServe(":8080", logRequest(http.DefaultServeMux)))
}
