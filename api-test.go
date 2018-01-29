package main

import "errors"

// TestProvider is a mock implementation for integration testing
type TestProvider struct {
	failing bool
}

// GetTranslation returns an error if failing flag is set, otherwise, simply returns the original string
func (p *TestProvider) GetTranslation(givenPhrase, givenLang, targetLang string, out *chan TranslateResult) {
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
}
