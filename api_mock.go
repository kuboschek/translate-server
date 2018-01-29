package main

import (
	"errors"
	"log"
	"time"
)

// TestProvider is a mock implementation for integration testing
type TestProvider struct {
	failing bool
	delay   time.Duration
}

// GetTranslation returns an error if failing flag is set, otherwise, simply returns the original string
func (p TestProvider) GetTranslation(givenPhrase, givenLang, targetLang string, out *chan TranslateResult) {
	log.Printf("TestProvider Got request: \"%v\" (%v -> %v)", givenPhrase, givenLang, targetLang)

	if p.delay > 0 {
		log.Printf("Simulating service delay of %v", p.delay)
		time.Sleep(p.delay)
	}

	if p.failing {
		*out <- TranslateResult{
			err: errors.New("simulating service failure"),
		}
	} else {
		*out <- TranslateResult{
			err:              nil,
			givenLang:        givenLang,
			givenPhrase:      givenPhrase,
			targetLang:       targetLang,
			translatedPhrase: givenPhrase,
		}
	}

	close(*out)
}
