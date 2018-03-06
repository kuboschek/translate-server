package upstream

import (
	"testing"
	"github.com/rubyist/circuitbreaker"
	"golang.org/x/text/language"
	"time"
)

const testPhrase = "TESTPHRASE"

func TestCircuitBreaker_TranslateWorking(t *testing.T) {
	breaker := circuit.NewBreaker()
	wrapper := CircuitBreaker{
		Breaker: breaker,
		Handler: Mock{},
	}
	out := make(chan Result)
	go wrapper.Translate(testPhrase, language.German, language.English, &out)

	select {
	case result := <- out:
		if result.Error != nil {
			t.Error("circuitbreaker returned error when it shouldn't have")
		}

		if result.TranslatedPhrase != testPhrase {
			t.Error("circuitbreaker returned incorrect result: want %v, got %v", testPhrase, result.TranslatedPhrase)
		}
		break;

	case <-time.After(time.Second):
		t.Error("operation timed out when it should have returned")
		break;
	}
}

func TestCircuitBreaker_TranslateTripped(t *testing.T) {
	breaker := circuit.NewBreaker()
	wrapper := CircuitBreaker{
		Breaker: breaker,
		Handler: Mock{},
	}

	// Trip the breaker, causing every request to error immediately
	breaker.Trip()

	out := make(chan Result)
	go wrapper.Translate(testPhrase, language.German, language.English, &out)

	select {
	case result := <-out:
		if result.Error == nil {
			t.Error("circuitbreaker returned no error when it should have")
		}
		break;

	case <-time.After(time.Second):
		t.Error("operation timed out when it should have returned")
		break;
	}

	breaker.Reset()
}

func TestCircuitBreaker_TranslateHandlerNil(t *testing.T) {
	breaker := circuit.NewBreaker()
	wrapper := CircuitBreaker{
		Breaker: breaker,
		Handler: nil,
	}
	out := make(chan Result)
	go wrapper.Translate(testPhrase, language.German, language.English, &out)

	select {
	case result := <- out:
		if result.Error == nil {
			t.Error("circuitbreaker returned no error when it should have")
		}
		break;

	case <-time.After(time.Second):
		t.Error("operation timed out when it should have returned")
		break;
	}
}

func TestCircuitBreaker_TranslateBreakerNil(t *testing.T) {
	wrapper := CircuitBreaker{
		Breaker: nil,
		Handler: Mock{},
	}
	out := make(chan Result)
	go wrapper.Translate(testPhrase, language.German, language.English, &out)

	select {
	case result := <- out:
		if result.Error == nil {
			t.Error("circuitbreaker returned no error when it should have")
		}
		break;

	case <-time.After(time.Second):
		t.Error("operation timed out when it should have returned")
		break;
	}
}

func TestCircuitBreaker_TranslateHandlerFail(t *testing.T) {
	breaker := circuit.NewBreaker()
	wrapper := CircuitBreaker{
		Breaker: breaker,
		Handler: Mock{Failing:true},
	}
	out := make(chan Result)
	go wrapper.Translate(testPhrase, language.German, language.English, &out)

	select {
	case result := <- out:
		if result.Error == nil {
			t.Error("circuitbreaker returned no error when it should have")
		}

		if breaker.Failures() != 1 {
			t.Error("circuitbreaker should record upstream errors")
		}
		break;

	case <-time.After(time.Second):
		t.Error("operation timed out when it should have returned")
		break;
	}
}