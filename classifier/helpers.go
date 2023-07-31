package classifier

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/navossoc/bayesian"
)

// check if file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// finds string in slice of strings
func sliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// checks if string is numeric
func isStringNumeric(word string) bool {
	return regexp.MustCompile(`^\d+$`).MatchString(word)
}

// checks if feature is not single character and not numeric
func validFeature(feature string) bool {
	return len(feature) > 1 && !isStringNumeric(feature)
}

// get slice of categories from training map
func getCategories(training map[string][]string) []bayesian.Class {
	var result []bayesian.Class
	for key := range training {
		result = append(result, bayesian.Class(key))
	}
	return result
}

// extract unique words from transaction description that are not numeric
func extractTransactionFeatures(transaction string) []string {
	var transFeatures []string
	features := strings.Split(transaction, " ")
	for _, feature := range features {
		if (len(feature) > 1) && (!sliceContains(transFeatures, feature)) && (!isStringNumeric(feature)) {
			transFeatures = append(transFeatures, feature)
		}
	}
	return transFeatures
}

// makes training data from CSV line entry
// returns category and slice of unique features
func parseTrainingEntry(data string) (string, []string) {
	splitLine := strings.Split(data, ",")
	//log.Printf("split transaction: %s\n", data)
	category := splitLine[0]
	words := strings.Split(splitLine[1], " ")
	var features []string
	for _, word := range words {
		if (validFeature(word)) && (!sliceContains(features, word)) {
			features = append(features, word)
		}
	}
	return category, features

}

// build training data map
// map[Category] = [feature1, feature2, ...]
func buildTrainingMap(data []string) map[string][]string {
	resultMap := make(map[string][]string)
	var features []string
	var category string
	for _, line := range data {
		category, features = parseTrainingEntry(line)
		log.Printf("training: %s - %s", category, features)
		_, exist := resultMap[category]
		if exist {
			resultMap[category] = append(resultMap[category], features...)
		} else {
			resultMap[category] = features
		}
	}
	return resultMap
}
