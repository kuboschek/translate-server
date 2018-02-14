package cache

import (
	"golang.org/x/text/language"
	"testing"
)

// Sample text constants for testing
const (
	testDe = "German"
	testEn = "English"
	testFr = "French"
	testSe = "Swedish"
)

var (
	de = language.German
	en = language.English
)

// TestAddTranslation tests the Put method of the memory cache
func TestAddTranslation(t *testing.T) {
	Memory.Put(testDe, en, testEn)
	Memory.Put(testDe, de, testFr)

	if Memory[testDe] == nil {
		t.Error("Put should create a map entry for every source phrase.")
	}

	if Memory[testDe][en] != testEn {
		t.Error("Put should create a map entry for every target phrase associated with a target language.")
	}
}

// TestGetTranslation tests the Get method of the memory cache
func TestGetTranslation(t *testing.T) {
	Memory[testFr] = map[language.Tag]string{
		de: testDe,
	}

	has := Memory.Has(testFr, de)
	if !has {
		t.Error("Has returned false, but translation is present.")
	}

	result, err := Memory.Get(testFr, de)
	if err != nil {
		t.Error("Get returned an error when a translation was present: %v", err)
	}

	if result != testDe {
		t.Errorf("Get returned the wrong translation: want %v got %v", testDe, result)
	}

	_, err = Memory.Get(testFr, en)
	if err == nil {
		t.Error("Get should return an error when no translation is present.")
	}

	_, err = Memory.Get(testSe, de)
	if err == nil {
		t.Error("Translate should return an error when no translation is present.")
	}
}
