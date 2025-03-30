package main

import (
	"ffiiitc/internal/classifier"
	"ffiiitc/internal/config"
	"ffiiitc/internal/firefly"
	"ffiiitc/internal/handlers"
	"ffiiitc/internal/router"

	"github.com/go-pkgz/lgr"
)

func main() {

	// make logger
	l := lgr.New(lgr.Debug, lgr.CallerFunc)
	l.Logf("INFO Firefly transaction classification started")

	// get the config
	l.Logf("INFO getting config")
	cfg, err := config.NewConfig(l)
	if err != nil {
		l.Logf("FATAL getting config: %v", err)
	}

	// make firefly http client for rest api
	fc := firefly.NewFireFlyHttpClient(cfg.FFApp, cfg.APIKey, config.FireflyAppTimeout, l)

	// make classifier
	// on first run, classifier will take all your
	// transactions and learn their categories
	// subsequent start classifier will load trained model from file
	l.Logf("INFO loading classifier from model: %s", config.ModelFile)
	cls, err := classifier.NewTrnClassifierFromFile(config.ModelFile, l)
	if err != nil {
		l.Logf("ERROR %v", err)
		l.Logf("INFO looks like we need to do some training...")
		// get transactions in data set
		//[ [cat, trn description], [cat, trn description]... ]
		trnDataset, err := fc.GetTransactionsDataset()
		l.Logf("DEBUG data set:\n %v", trnDataset)

		if err != nil {
			l.Logf("FATAL: unable to get list of transactions %v", err)
		}

		// byesian package requires at least 2 transactions with different categories to start training

		// we fail if no transactions in Firefly
		if len(trnDataset) == 0 {
			l.Logf("FATAL: no transactions in Firefly. At least 2 manually categorised transactions with different categories are required %v", trnDataset)
		}

		// we also check for at least 2 different categories available if transactions exist
		categories := make(map[string]int)
		for i, data := range trnDataset {
			if len(data) == 0 {
				l.Logf("WARN skipping empty transaction data at index %d", i)
				continue
			}
			category := data[0]
			if category == "" {
				l.Logf("WARN skipping transaction with empty category at index %d", i)
				continue
			}
			categories[category]++
		}

		l.Logf("INFO found %d different categories: %v", len(categories), categories)

		if len(categories) < 2 {
			l.Logf("FATAL: classifier needs at least 2 different categories in transactions, got %d. Categories found: %v",
				len(categories),
				categories)
			return
		}

		cls, err = classifier.NewTrnClassifierWithTraining(trnDataset, l)
		if err != nil {
			l.Logf("FATAL: %v", err)
		}
		l.Logf("INFO training completed...")
		err = cls.SaveClassifierToFile(config.ModelFile)
		if err != nil {
			l.Logf("FATAL: %v", err)
		}
		l.Logf("INFO classifier saved to: %s", config.ModelFile)
	}

	l.Logf("DEBUG learned classes: %v", cls.Classifier.Classes)

	// init handlers
	h := handlers.NewWebHookHandler(cls, fc, l)

	// init router
	r := router.NewRouter()

	// add handlers
	r.AddRoute("/classify", h.HandleNewTransactionWebHook)
	r.AddRoute("/train", h.HandleForceTrainingModel)
	// temporary remove this handle
	//r.AddRoute("/learn", h.HandleUpdateTransactionWebHook)

	//run
	err = r.Run(8080)
	if err != nil {
		panic(err)
	}
}
