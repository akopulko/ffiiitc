package handlers

import (
	"encoding/json"
	"ffiiitc/internal/classifier"
	"ffiiitc/internal/firefly"
	"net/http"
	"strconv"

	"github.com/go-pkgz/lgr"
	"github.com/navossoc/bayesian"
)

type WebHookHandler struct {
	Classifier    *classifier.TrnClassifier
	FireflyClient *firefly.FireFlyHttpClient
	Logger        *lgr.Logger
}

// structs to handle payload from new transaction web hook
type FireflyTrn struct {
	Id          int64  `json:"transaction_journal_id"`
	Description string `json:"description"`
	Category    string `json:"category_name"`
}

type FireFlyContent struct {
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
			trn.Id,
			trn.Description,
		)
		cat := wh.Classifier.ClassifyTransaction(trn.Description)
		wh.Logger.Logf("INFO hook new trn: classified (id: %v) (category: %s)", trn.Id, cat)
		err = wh.FireflyClient.UpdateTransactionCategory(strconv.FormatInt(trn.Id, 10), cat)
		if err != nil {
			wh.Logger.Logf("ERROR hook new trn: error updating (id: %v) %v", trn.Id, err)
		}
		wh.Logger.Logf("INFO hook new trn: updated (id: %v)", trn.Id)

	}
	w.WriteHeader(http.StatusOK)
}

// http handler for new transaction
func (wh *WebHookHandler) HandleUpdateTransactionWebHook(w http.ResponseWriter, r *http.Request) {

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

	// perform training
	for _, trn := range hookData.Content.Transactions {
		wh.Logger.Logf(
			"hook update trn: received (id: %v) (desc: %s) (cat: %s)",
			trn.Id,
			trn.Description,
			trn.Category,
		)

		if trn.Category != "" {
			err := wh.Classifier.Train(trn.Description, trn.Category)
			if err != nil {
				wh.Logger.Logf("hook update trn: error updating model: %v", err)
			}
			wh.Logger.Logf(
				"hook update trn: (cat: %s) (features: %v)",
				trn.Category,
				wh.Classifier.Classifier.WordsByClass(bayesian.Class(trn.Category)),
			)
		} else {
			wh.Logger.Logf("skip training. Category is empty")
		}

	}
	w.WriteHeader(http.StatusOK)
}
