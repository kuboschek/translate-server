package main

// TranslateResult represents the result of a translation call to a service
type TranslateResult struct {
	err              error
	givenPhrase      string
	givenLang        string
	targetLang       string
	translatedPhrase string
}

// TranslateService represents an external service that provides translations
type TranslateService interface {
	GetTranslation(givenPhrase, givenLang, targetLang string, out *chan TranslateResult)
}

// TODO(lkuboschek): Write Mock API
