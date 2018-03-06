package upstream

import (
	"github.com/rubyist/circuitbreaker"
	"golang.org/x/text/language"
	"log"
	"time"
	"errors"
)

// CircuitBreaker implements a wrapper for upstream handlers that uses a CircuitBreaker to
// short-circuit requests given a user-specified circuit breaker
type CircuitBreaker struct {
	Breaker *circuit.Breaker
	Timeout time.Duration
	Handler Service
}

func (b *CircuitBreaker) Translate (givenPhrase string, givenLang, targetLang language.Tag, out *chan Result) {
	defer close(*out)
	if b.Breaker == nil {
		*out <- Result{
			Error: errors.New("circuit breaker: is nil"),
		}
		return
	}

	if b.Breaker.Tripped() {
		*out <- Result{
			Error: errors.New("circuit breaker: is tripped"),
		}
		return
	}

	intermediate := make(chan Result)
	if(b.Handler != nil) {
		go b.Handler.Translate(givenPhrase, givenLang, targetLang, &intermediate)
		for result := range intermediate {
			if result.Error != nil {
				b.Breaker.Fail()
			}

			*out <- result
			return
		}
	} else {
		log.Print("circuit breaker: wrapped handler is nil")
		*out <- Result{
			Error: errors.New("wrapped handler is nil"),
		}
		b.Breaker.Fail()
	}
}