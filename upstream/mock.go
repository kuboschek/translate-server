package upstream

import (
	"errors"
	"golang.org/x/text/language"
	"log"
	"time"
)

// Mock is an implementation used to test error handling behaviour.
type Mock struct {
	Failing bool
	Delay   time.Duration
}

// Translate returns an error if Failing flag is set. Otherwise, simply returns the original string.
// If Delay is set to non-zero values, waits for the given time before responding.
func (p Mock) Translate(givenPhrase string, givenLang, targetLang language.Tag, out *chan Result) {
	log.Printf("Mock (%#v) got request: \"%v\" (%v -> %v)", p, givenPhrase, givenLang, targetLang)

	if p.Delay > 0 {
		log.Printf("Simulating service delay of %v", p.Delay)
		time.Sleep(p.Delay)
	}

	if p.Failing {
		log.Println("Simulating service failure.")
		*out <- Result{
			Error: errors.New("simulating service failure"),
		}
	} else {
		*out <- Result{
			Error:            nil,
			GivenLang:        givenLang,
			GivenPhrase:      givenPhrase,
			TargetLang:       targetLang,
			TranslatedPhrase: givenPhrase,
		}
	}
}
