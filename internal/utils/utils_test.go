package utils

import (
	"testing"

	"github.com/navossoc/bayesian"
	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	existingFile := "utils.go"
	nonExistingFile := "blah.txt"

	assert.True(t, FileExists(existingFile))
	assert.False(t, FileExists(nonExistingFile))
}

func TestSliceContains(t *testing.T) {
	slice := []string{"apple", "banana", "orange"}

	assert.True(t, sliceContains(slice, "apple"))
	assert.True(t, sliceContains(slice, "banana"))
	assert.True(t, sliceContains(slice, "orange"))
	assert.False(t, sliceContains(slice, "grape"))
}

func TestIsStringNumeric(t *testing.T) {
	assert.True(t, isStringNumeric("123"))
	assert.True(t, isStringNumeric("4567890"))
	assert.False(t, isStringNumeric("abc"))
	assert.False(t, isStringNumeric("123abc"))
}

func TestValidFeature(t *testing.T) {
	assert.True(t, validFeature("apple"))
	assert.True(t, validFeature("banana"))
	assert.False(t, validFeature("a"))
	assert.False(t, validFeature("123"))
}

func TestGetCategories(t *testing.T) {
	trainingData := map[string][]string{
		"fruit":  {"apple", "banana", "orange"},
		"animal": {"dog", "cat", "bird"},
	}

	categories := GetCategories(trainingData)

	assert.ElementsMatch(t, categories, []bayesian.Class{"fruit", "animal"})
}

func TestExtractTransactionFeatures(t *testing.T) {
	transaction := "buying apples and oranges at the store"
	expectedFeatures := []string{"buying", "apples", "and", "oranges", "at", "the", "store"}

	features := ExtractTransactionFeatures(transaction)

	assert.ElementsMatch(t, features, expectedFeatures)
}

func TestParseTrainingEntry(t *testing.T) {
	data := "fruit,apple banana orange apple"

	category, features := parseTrainingEntry(data)

	assert.Equal(t, category, "fruit")
	assert.ElementsMatch(t, features, []string{"apple", "banana", "orange"})
}

func TestBuildTrainingMap(t *testing.T) {
	data := []string{
		"fruit,apple banana orange",
		"animal,dog cat bird",
		"fruit,grape apple",
	}

	expectedMap := map[string][]string{
		"fruit":  {"apple", "banana", "orange", "grape", "apple"},
		"animal": {"dog", "cat", "bird"},
	}

	trainingMap := BuildTrainingMap(data)

	assert.Equal(t, expectedMap, trainingMap)
}
