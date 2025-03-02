package handlers

import (
	"encoding/json"
	"ffiiitc/internal/classifier"
	"ffiiitc/internal/config"
	"ffiiitc/internal/firefly"

	"net/http"

	"github.com/go-pkgz/lgr"
)

type WebHookHandler struct {
	Classifier    *classifier.TrnClassifier
	FireflyClient *firefly.FireFlyHttpClient
	Logger        *lgr.Logger
}

// structs to handle payload from new transaction web hook
type FireflyTrn struct {
	Id          string  `json:"transaction_journal_id"`
	Description string `json:"description"`
	Category    string `json:"category_name"`
}

type FireFlyContent struct {
	Id           string        `json:"id"`
	Transactions []FireflyTrn `json:"transactions"`
}

type FireflyWebHook struct {
	Content FireFlyContent `json:"content"`
}

func NewWebHookHandler(c *classifier.TrnClassifier, f *firefly.FireFlyHttpClient, l *lgr.Logger) *WebHookHandler {
	return &WebHookHandler{
		Classifier:    c,
		FireflyClient: f,
		Logger:        l,
	}
}

// http handler for new transaction
func (wh *WebHookHandler) HandleNewTransactionWebHook(w http.ResponseWriter, r *http.Request) {

	// only allow post method
	if r.Method != http.MethodPost {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// decode payload
	decoder := json.NewDecoder(r.Body)
	var hookData FireflyWebHook
	err := decoder.Decode(&hookData)
	if err != nil {
		http.Error(w, "bad data", http.StatusBadRequest)
		return
	}

	// perform classification
	for _, trn := range hookData.Content.Transactions {
		wh.Logger.Logf(
			"INFO hook new trn: received (id: %v) (description: %s)",
			hookData.Content.Id,
			trn.Description,
		)
		cat := wh.Classifier.ClassifyTransaction(trn.Description)
		wh.Logger.Logf("INFO hook new trn: classified (id: %v) (category: %s)", hookData.Content.Id, cat)
		err = wh.FireflyClient.UpdateTransactionCategory(hookData.Content.Id, trn.Id, cat)
		if err != nil {
			wh.Logger.Logf("ERROR hook new trn: error updating (id: %v) %v", hookData.Content.Id, err)
		}
		wh.Logger.Logf("INFO hook new trn: updated (id: %v)", hookData.Content.Id)

	}
	w.WriteHeader(http.StatusOK)
}

// http handler for forcing to train model
func (wh *WebHookHandler) HandleForceTrainingModel(w http.ResponseWriter, r *http.Request) {

	// only allow post method
	if r.Method != http.MethodGet {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	wh.Logger.Logf("INFO Received request to perform force training")
	wh.Logger.Logf("INFO Requesting transactions data from Firefly")
	trnDataset, err := wh.FireflyClient.GetTransactionsDataset()
	if err != nil || len(trnDataset) == 0 {
		wh.Logger.Logf("ERROR: Error while getting transactions data\n %v", err)
	} else {
		wh.Logger.Logf("DEBUG Got training data\n %v", trnDataset)
		cls, err := classifier.NewTrnClassifierWithTraining(trnDataset, wh.Logger)
		if err != nil {
			wh.Logger.Logf("ERROR creating classifier from dataset:\n %v", err)
		} else {
			wh.Logger.Logf("INFO forced training completed...")
			wh.Logger.Logf("INFO saving data to model...")
			err = cls.SaveClassifierToFile(config.ModelFile)
			if err != nil {
				wh.Logger.Logf("ERROR saving model to file:\n %v", err)
			} else {
				wh.Logger.Logf("INFO: forced training completed and model saved. Please restart 'ffiiitc' container")
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// http handler for updated transaction
// func (wh *WebHookHandler) HandleUpdateTransactionWebHook(w http.ResponseWriter, r *http.Request) {

// 	// only allow post method
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "bad request", http.StatusBadRequest)
// 		return
// 	}

// 	// decode payload
// 	decoder := json.NewDecoder(r.Body)
// 	var hookData FireflyWebHook
// 	err := decoder.Decode(&hookData)
// 	if err != nil {
// 		http.Error(w, "bad data", http.StatusBadRequest)
// 		return
// 	}

// 	// perform training
// 	for _, trn := range hookData.Content.Transactions {
// 		wh.Logger.Logf(
// 			"hook update trn: received (id: %v) (desc: %s) (cat: %s)",
// 			trn.Id,
// 			trn.Description,
// 			trn.Category,
// 		)

// 		if trn.Category != "" {
// 			err := wh.Classifier.Train(trn.Description, trn.Category)
// 			if err != nil {
// 				wh.Logger.Logf("hook update trn: error updating model: %v", err)
// 			}
// 			wh.Logger.Logf(
// 				"hook update trn: (cat: %s) (features: %v)",
// 				trn.Category,
// 				wh.Classifier.Classifier.WordsByClass(bayesian.Class(trn.Category)),
// 			)
// 		} else {
// 			wh.Logger.Logf("skip training. Category is empty")
// 		}

// 	// }
// 	w.WriteHeader(http.StatusOK)
// }
