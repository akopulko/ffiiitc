package classifier

import (
	"regexp"
	"slices"
	"strings"

	"github.com/go-pkgz/lgr"
	"github.com/navossoc/bayesian"
)

// classifier implementation
type TrnClassifier struct {
	Classifier *bayesian.Classifier
	logger     *lgr.Logger
}

type TransactionDataSet [][]string

// init classifier with model file
func NewTrnClassifierFromFile(modelFile string, l *lgr.Logger) (*TrnClassifier, error) {
	cls, err := bayesian.NewClassifierFromFile(modelFile)
	if err != nil {
		return nil, err
	}
	return &TrnClassifier{
		Classifier: cls,
		logger:     l,
	}, nil
}

// init classifier with training data set
func NewTrnClassifierWithTraining(dataSet TransactionDataSet, l *lgr.Logger) (*TrnClassifier, error) {
	trainingMap := convertDatasetToTrainingMap(dataSet)
	catList := getCategoriesFromTrainingMap(trainingMap)
	//catList := maps.Keys(trainingMap)
	cls := bayesian.NewClassifier(catList...)
	for _, cat := range catList {
		cls.Learn(trainingMap[string(cat)], cat)
	}
	return &TrnClassifier{
		Classifier: cls,
		logger:     l,
	}, nil
}

// save classifier to model file
func (tc *TrnClassifier) SaveClassifierToFile(modelFile string) error {
	err := tc.Classifier.WriteToFile(modelFile)
	return err
}

// perform transaction classification
// in: transaction description
// out: likely transaction category
func (tc *TrnClassifier) ClassifyTransaction(t string) string {
	features := extractTransactionFeatures(t)
	_, likely, _ := tc.Classifier.LogScores(features)
	return string(tc.Classifier.Classes[likely])
}

// function to get category and list of
// unique features from line of transaction data set
// in: [cat, trn description]
// out: cat, [features...]
func getCategoryAndFeatures(data []string) (string, []string) {
	category := data[0]
	words := strings.Split(data[1], " ")
	var features []string
	for _, word := range words {
		if (validFeature(word)) && (!slices.Contains(features, word)) {
			features = append(features, word)
		}
	}
	return category, features

}

// get slice of categories from training map
func getCategoriesFromTrainingMap(training map[string][]string) []bayesian.Class {
	var result []bayesian.Class
	for key := range training {
		result = append(result, bayesian.Class(key))
	}
	return result
}

// checks if feature is valid
// should be not single symbol and not pure number
func validFeature(feature string) bool {
	return len(feature) > 1 && !isStringNumeric(feature)
}

// checks if string is pure number: int or float
func isStringNumeric(s string) bool {
	numericPattern := `^-?\d+(\.\d+)?$`
	match, err := regexp.MatchString(numericPattern, s)
	return err == nil && match
}

// build training map from transactions data set
// in: [ [cat, trn description], [cat, trn description]... ]
// out: map[Category] = [feature1, feature2, ...]
func convertDatasetToTrainingMap(dataSet TransactionDataSet) map[string][]string {
	resultMap := make(map[string][]string)
	var features []string
	var category string
	for _, line := range dataSet {
		category, features = getCategoryAndFeatures(line)
		_, exist := resultMap[category]
		if exist {
			resultMap[category] = append(resultMap[category], features...)
		} else {
			resultMap[category] = features
		}
	}
	return resultMap
}

// extract unique words from transaction description that are not numeric
func extractTransactionFeatures(transaction string) []string {
	var transFeatures []string
	features := strings.Split(transaction, " ")
	for _, feature := range features {
		if (len(feature) > 1) && (!slices.Contains(transFeatures, feature)) && (!isStringNumeric(feature)) {
			transFeatures = append(transFeatures, feature)
		}
	}
	return transFeatures
}
