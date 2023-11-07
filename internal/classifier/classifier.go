package classifier

import (
	"ffiiitc/internal/firefly"
	"ffiiitc/internal/utils"
	"fmt"
	"os"
    "time"
	"github.com/go-pkgz/lgr"
	"github.com/navossoc/bayesian"
)

const (
	modelFile = "data/model.gob" //file name to store model
)

// classifier implementation
type TrnClassifier struct {
	Classifier *bayesian.Classifier
	logger     *lgr.Logger
}

// when creating new classifier, it will first try to load trained data
// from file. If not possible, it will then get all you existing categories transactions from
// Firefly and train itself on this data set
func NewTrnClassifier(fc *firefly.FireFlyHttpClient, l *lgr.Logger) *TrnClassifier {

	l.Logf("INFO trying to load classifier from %s", modelFile)
	c, err := loadClassifierFromFile(modelFile)
	file, _ := os.Stat(modelFile)
	if err != nil || time.Now().Sub(file.ModTime()) > 12*time.Hour{
		l.Logf("ERROR loading classifier from file %s, %v", modelFile, err)
		//log.Println("no model file found, need to do some training...")
		l.Logf("INFO need to do some training...")
		trn, err := fc.GetTransactions()
		if err != nil {
			l.Logf("FATAL unable to get transactions from firefly: %v", err)

		}
		cls, err := trainClassifierFromTransactions(trn)
		if err != nil {
			l.Logf("FATAL train classifier: %v", err)
		}
		l.Logf("INFO training completed")
		err = saveClassifierToFile(cls, modelFile)
		if err != nil {
			l.Logf("FATAL save classifier to file %s: %v", modelFile, err)
		}
		l.Logf("INFO trained model successfully saved to: %s for future use\n", modelFile)
		return &TrnClassifier{
			Classifier: cls,
			logger:     l,
		}
	}
	return &TrnClassifier{
		Classifier: c,
		logger:     l,
	}
}

func (tc *TrnClassifier) ClassifyTransaction(t string) string {
	features := utils.ExtractTransactionFeatures(t)
	_, likely, _ := tc.Classifier.LogScores(features)
	return string(tc.Classifier.Classes[likely])
}

func (tc *TrnClassifier) Train(transaction string, category string) error {
	features := utils.ExtractTransactionFeatures(transaction)
	tc.Classifier.Learn(features, bayesian.Class(category))
	err := tc.Classifier.WriteToFile(modelFile)
	return err
}

func loadClassifierFromFile(modelFile string) (*bayesian.Classifier, error) {
	if utils.FileExists(modelFile) {
		cls, err := bayesian.NewClassifierFromFile(modelFile)
		if err != nil {
			return nil, err
		}
		return cls, nil
	}
	return nil, fmt.Errorf("model file does not exist")
}

func saveClassifierToFile(c *bayesian.Classifier, modelFile string) error {
	err := c.WriteToFile(modelFile)
	return err
}

func trainClassifierFromTransactions(transactions []string) (*bayesian.Classifier, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions provided for training")
	}

	trainingMap := utils.BuildTrainingMap(transactions)
	cat := utils.GetCategories(trainingMap)
	cls := bayesian.NewClassifier(cat...)
	for _, categ := range cat {
		cls.Learn(trainingMap[string(categ)], categ)
	}

	return cls, nil
}
