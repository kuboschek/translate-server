package main

import (
	"bytes"
	"testing"
)

const (
	testDe = "testde"
	testEn = "testen"
	testFr = "testfr"
	testSe = "testse"

	de = "de"
	en = "en"
	fr = "fr"
)

func TestAddTranslation(t *testing.T) {
	AddTranslation(en, testDe, testEn)
	AddTranslation(fr, testDe, testFr)

	if phraseMap[testDe] == nil {
		t.Error("AddTranslation should create a map entry for every source phrase.")
	}

	if phraseMap[testDe][en] != testEn {
		t.Error("AddTranslation should create a map entry for every target phrase associated with a target language.")
	}
}

func TestGetTranslation(t *testing.T) {
	phraseMap[testFr] = map[string]string{
		de: testDe,
	}

	trans1 := GetTranslation(de, testFr)
	if trans1 == nil {
		t.Error("GetTranslation returned nil when a translation was present.")
	}

	if *trans1 != testDe {
		t.Errorf("GetTranslation returned the wrong translation: want %v got %v", testDe, trans1)
	}

	trans2 := GetTranslation(en, testFr)
	if trans2 != nil {
		t.Error("GetTranslation should return nil when no translation is present.")
	}

	trans3 := GetTranslation(de, testSe)
	if trans3 != nil {
		t.Error("GetTranslation should return nil when no translation is present.")
	}
}

func TestStoreLoad(t *testing.T) {
	buf := new(bytes.Buffer)

	StoreCache(buf)
	LoadCache(buf)
}
