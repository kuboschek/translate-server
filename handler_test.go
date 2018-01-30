package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const (
	testContent = "testContent"
	testLang    = "testLang"
)

func TestWriteResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	writeResponse(rr, testLang, testContent)

	if rr.Body.String() != testContent {
		t.Error("writeResponse should write exactly the given phrase to the output.")
	}

	if rr.Header().Get("Content-Language") != testLang {
		t.Error("writeResponse should set the Content-Language header.")
	}

	if rr.Header().Get("Content-Type") != "text/plain" {
		t.Error("writeResponse should set the Content-Type header.")
	}
}
func TestTranslateHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("translateHandler should not allow %v requests.", http.MethodGet)
	}
}

func TestTranslateHandler2(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Language", "en;q=1.2.3")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Error("translateHandler should reject malformed Accept-Language headers.")
	}
}

func TestTranslateHandler3(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Language", "en,fr")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject request with missing Content-Language headers.")
	}
}

func TestTranslateHandler4(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Language", "fr")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject request with missing Accept-Language headers.")
	}
}

func TestTranslateHandler5(t *testing.T) {
	content := bytes.NewBufferString("Guten Morgen.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{
		TestProvider{},
	}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("translateHandler should accept requests with appropriate headers.")
	}
}

func TestTranslateHandler6(t *testing.T) {
	content := bytes.NewBufferString("Whatever.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("translateHandler should reject requests when all services fail: got %v want %v", rr.Code, http.StatusInternalServerError)
	}
}

const cacheString = "This is a different test."

func TestTranslateHandler7(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	content := bytes.NewBufferString(cacheString)
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler = append(translateHandler,
		TestProvider{
			failing: true,
			delay:   time.Nanosecond,
		},
	)
	translateHandler.ServeHTTP(rr, req)

	fmt.Println(buf.String())
	if buf.Len() == 0 {
		t.Error("translateHandler should log when a service fails.")
	}
}

// TestTranslateHandler8 tests if caching returns correct results
func TestTranslateHandler8(t *testing.T) {
	content := bytes.NewBufferString(cacheString)
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}

	// This fills the cache
	translateHandler = append(translateHandler,
		TestProvider{
			failing: false,
			delay:   time.Nanosecond,
		},
	)
	translateHandler.ServeHTTP(rr, req)

	// Empty the services list, so any cache miss will result in an error
	translateHandler = TranslateHandler{}
	rr = httptest.NewRecorder()
	content = bytes.NewBufferString(cacheString)
	req = httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")

	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected OK status: got %v want %v", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != cacheString {
		t.Errorf("caching should return previous results: got %v want %v", rr.Body.String(), cacheString)
	}
}

func TestMoveToBack(t *testing.T) {
	var services = TranslateHandler{
		TestProvider{
			failing: true,
		},
		TestProvider{
			delay: time.Second * 123,
		},
	}

	services.moveToBack(0)

	if len(services) != 2 {
		t.Errorf("moveToBack should not remove or add items. got len %v want len %v", len(services), 2)
	}

	if services[0].(TestProvider).delay != time.Second*123 {
		t.Error("moveToBack should move the service with a given index to the back of the list.")
	}

	if services[1].(TestProvider).failing != true {
		t.Error("moveToBack should move the specified service to the back of the list.")
	}
}
