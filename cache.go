package main

import (
	"encoding/json"
	"io"
)

var phraseMap map[string]map[string]string

func init() {
	phraseMap = make(map[string]map[string]string)
}

// AddTranslation stores a translation in the memory cache
func AddTranslation(targetLang, sourcePhrase, targetPhrase string) {

	if phraseMap[sourcePhrase] == nil {
		phraseMap[sourcePhrase] = make(map[string]string)
	}

	phraseMap[sourcePhrase][targetLang] = targetPhrase
}

// GetTranslation returns a translation from cache, or nil, if it's not present
func GetTranslation(targetLang, sourcePhrase string) (targetPhrase *string) {
	phrases, ok := phraseMap[sourcePhrase]
	if !ok {
		return nil
	}

	phrase, ok := phrases[targetLang]
	if !ok {
		return nil
	}

	return &phrase
}

// StoreCache writes the current phrase cache to the writer.
func StoreCache(writer io.Writer) error {
	enc := json.NewEncoder(writer)
	return enc.Encode(phraseMap)
}

// LoadCache reads the phrase cache from the reader.
func LoadCache(reader io.Reader) error {
	dec := json.NewDecoder(reader)
	return dec.Decode(&phraseMap)
}
