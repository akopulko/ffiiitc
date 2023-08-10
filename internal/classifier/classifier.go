package classifier

import (
	"ffiiitc/internal/firefly"
	"ffiiitc/internal/utils"
	"fmt"
	"log"

	"github.com/go-pkgz/lgr"
	"github.com/navossoc/bayesian"
)

const (
	modelFile = "data/model.gob"
)

type TrnClassifier struct {
	Classifier *bayesian.Classifier
	logger     *lgr.Logger
}

func NewTrnClassifier(fc *firefly.FireFlyHttpClient, l *lgr.Logger) *TrnClassifier {

	c, err := loadClassifierFromFile(modelFile)
	if err != nil {
		l.Logf("ERROR loading classifier from file %s, %v", modelFile, err)
		//log.Println("no model file found, need to do some training...")
		l.Logf("INFO need to do some training...")
		trn, err := fc.GetTransactions()
		if err != nil {
			log.Println(err)
		}
		cls, err := trainClassifierFromTransactions(trn)
		if err != nil {
			log.Fatal(err)
		}
		//log.Println("training completed")
		l.Logf("INFO training completed")
		err = saveClassifierToFile(cls, modelFile)
		if err != nil {
			log.Fatal(err)
		}
		//log.Printf("trained model successfully saved to: %s for future use\n", modelFile)
		l.Logf("INFO trained model successfully saved to: %s for future use\n", modelFile)
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

// func loadClassifierFromFile(modelFile string) (*bayesian.Classifier, error) {
// 	if utils.FileExists(modelFile) {
// 		log.Printf("loading model from file: %s", modelFile)
// 		cls, err := bayesian.NewClassifierFromFile(modelFile)
// 		if err != nil {
// 			log.Printf("error loading from file: %s", modelFile)
// 			log.Fatal(err)
// 		}
// 		return cls, nil
// 	}
// 	return nil, fmt.Errorf("model file does not exist")
// }

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
