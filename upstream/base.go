
// Package upstream provides implementations for upstream translation services.
package upstream

import "golang.org/x/text/language"

// Result represents the result of a translation call to a service.
type Result struct {
	Error            error
	GivenPhrase      string
	GivenLang        language.Tag
	TargetLang       language.Tag
	TranslatedPhrase string
}

// Service represents an external service that provides translations.
type Service interface {
	Translate(givenPhrase string, givenLang, targetLang language.Tag, out *chan Result)
}
