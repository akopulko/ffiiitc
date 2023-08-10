package main

import (
	"ffiiitc/internal/classifier"
	"ffiiitc/internal/config"
	"ffiiitc/internal/firefly"
	"ffiiitc/internal/handlers"
	"ffiiitc/internal/router"

	"github.com/go-pkgz/lgr"
)

const (
	ffAppTimeout = 10 // 10 sec for fftc to app service timeout
)

func main() {

	// make logger
	l := lgr.New(lgr.Debug, lgr.CallerFunc)
	l.Logf("INFO Firefly transaction classification started")

	// get the config
	l.Logf("INFO getting config")
	cfg, err := config.NewConfig()
	if err != nil {
		l.Logf("FATAL getting config: %v", err)
	}

	// make firefly http client for rest api
	fc := firefly.NewFireFlyHttpClient(cfg.FFApp, cfg.APIKey, ffAppTimeout, l)

	// make classifier
	// on first run, classifier will take all your
	// transactions and learn their categories
	// subsequent start classifier will load trained model from file
	cls := classifier.NewTrnClassifier(fc, l)
	l.Logf("INFO learned classes: %v", cls.Classifier.Classes)

	// init handlers
	h := handlers.NewWebHookHandler(cls, fc, l)

	// init router
	r := router.NewRouter()

	// add handlers
	r.AddRoute("/classify", h.HandleNewTransactionWebHook)
	r.AddRoute("/learn", h.HandleUpdateTransactionWebHook)

	//run
	err = r.Run(8080)
	if err != nil {
		panic(err)
	}
}
