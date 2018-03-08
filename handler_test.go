package main

import (
	"bytes"
	"fmt"
	"github.com/kuboschek/translate-server/cache"
	"github.com/kuboschek/translate-server/upstream"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const (
	testContent = "testContent"
)

var (
	testLang = language.Und
)

func TestWriteResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	writeSuccess(rr, testLang, testContent)

	if rr.Body.String() != testContent {
		t.Error("writeSuccess should write exactly the given phrase to the output.")
	}

	if rr.Header().Get("Content-Language") != testLang.String() {
		t.Error("writeSuccess should set the Content-Language header.")
	}

	if rr.Header().Get("Content-Type") != "text/plain" {
		t.Error("writeSuccess should set the Content-Type header.")
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

// TestRejectMalformedAcceptLanguage checks that the handler rejects malformed Accept-Language headers
func TestRejectMalformedAcceptLanguage(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Language", "en;q=1.2.3")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject malformed Accept-Language headers.")
	}
}

// TestRejectMissingContentLanguage checks that the handler requires a Content-Language header
func TestRejectMissingContentLanguage(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Language", "en,fr")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject request with missing Content-Language headers.")
	}
}

// TestRequireAcceptLanguageHeader checks that the handler requires an Accept-Language header
func TestRequireAcceptLanguageHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Language", "fr")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject request with missing Accept-Language headers.")
	}
}

// TestSuccessResponse checks if well-formed requests return correct results
// when using a functioning upstream
func TestSuccessResponse(t *testing.T) {
	content := bytes.NewBufferString("Guten Morgen.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{
		[]upstream.Service{
			upstream.Mock{},
		},
		nil,
	}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("translateHandler should accept requests with appropriate headers.")
	}
}

// TestAllUpstreamFailure checks if a 502 is returned when all upstreams fail
func TestAllUpstreamFailure(t *testing.T) {
	content := bytes.NewBufferString("Whatever.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("translateHandler should reject requests when all services fail: got %v want %v", rr.Code, http.StatusBadGateway)
	}
}

const cacheString = "This is a different test."

// TestLogOnFailure checks if the handler logs on upstream errors
func TestLogOnFailure(t *testing.T) {
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

	handler := TranslateHandler{}
	handler.Services = append(handler.Services,
		upstream.Mock{
			Failing: true,
			Delay:   time.Nanosecond,
		},
	)
	handler.ServeHTTP(rr, req)

	fmt.Println(buf.String())
	if buf.Len() == 0 {
		t.Error("TranslateHandler should log when a service fails.")
	}
}

// TestCachedResponseCorrect tests if caching returns correct results
func TestCachedResponseCorrect(t *testing.T) {
	content := bytes.NewBufferString(cacheString)
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	handler := TranslateHandler{
		[]upstream.Service{
			upstream.Mock{
				Failing: false,
				Delay:   0,
			},
		},
		cache.Memory,
	}
	handler.ServeHTTP(rr, req)

	// Empty the services list, so any cache miss will result in an error
	handler = TranslateHandler{
		Cache: cache.Memory,
	}
	rr = httptest.NewRecorder()
	content = bytes.NewBufferString(cacheString)
	req = httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")

	// Wait here to allow the async cache write to finish
	time.Sleep(time.Second)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected OK status: got %v want %v", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != cacheString {
		t.Errorf("caching should return previous results: got %v want %v", rr.Body.String(), cacheString)
	}
}

// TestTimeOut checks if the handler times out after a certain time
func TestTimeOut(t *testing.T) {
	content := bytes.NewBufferString(cacheString)
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	var handler = TranslateHandler{
		[]upstream.Service{
			upstream.Mock{
				Delay: time.Second * 123,
			},
		},
		nil,
	}

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Error("handler should return a 502 when upstream services time out")
	}
}

// TestMultipleUpstreams checks if well-formed requests return correct results
// when using multiple upstreams
func TestMultipleUpstreams(t *testing.T) {
	content := bytes.NewBufferString("Guten Morgen.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler := TranslateHandler{
		[]upstream.Service{
			upstream.Mock{Failing: true},
			upstream.Mock{},
		},
		nil,
	}
	translateHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("translateHandler should accept requests with appropriate headers.")
	}
}
