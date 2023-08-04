package classifier

import (
	"ffiiitc/firefly"
	"log"

	"github.com/navossoc/bayesian"
)

const modelFile = "data/model.gob"

type TrnClassifier struct {
	Classifier    *bayesian.Classifier
	FireFlyClient *firefly.FireFlyHttpClient
}

func NewTrnClassifier(fc *firefly.FireFlyHttpClient) *TrnClassifier {
	log.Printf("looking for model file: %s", modelFile)
	if fileExists(modelFile) {
		log.Println("loading model from file")
		cls, err := bayesian.NewClassifierFromFile(modelFile)
		if err != nil {
			log.Fatal(err)
		}
		return &TrnClassifier{
			Classifier:    cls,
			FireFlyClient: fc,
		}
	} else {
		log.Println("no model file found, need to do some training...")
		trn, err := fc.GetTransactions()
		if err != nil {
			log.Println(err)
		}
		if len(trn) == 0 {
			log.Fatal("no transactions found in FireFly. You need to have at least some of transaction categorised for system to train.")
		}
		trainingMap := buildTrainingMap(trn)
		cat := getCategories(trainingMap)
		cls := bayesian.NewClassifier(cat...)
		for _, categ := range cat {
			cls.Learn(trainingMap[string(categ)], categ)
		}
		log.Println("training completed")
		err = cls.WriteToFile(modelFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("trained model successfully saved to: %s for future use\n", modelFile)
		return &TrnClassifier{
			Classifier:    cls,
			FireFlyClient: fc,
		}
	}
}

func (tc *TrnClassifier) ClassifyTransaction(t string) string {
	features := extractTransactionFeatures(t)
	_, likely, _ := tc.Classifier.LogScores(features)
	return string(tc.Classifier.Classes[likely])
}

func (tc *TrnClassifier) Train(transaction string, category string) error {
	features := extractTransactionFeatures(transaction)
	tc.Classifier.Learn(features, bayesian.Class(category))
	err := tc.Classifier.WriteToFile(modelFile)
	return err
}
