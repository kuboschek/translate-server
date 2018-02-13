package upstream

import (
	"golang.org/x/text/language"
	"net/url"
	"net/http"
	"bytes"
	"github.com/pkg/errors"
	"log"
	"encoding/xml"
)

const (
	bingAPIBase = "https://api.microsofttranslator.com/v2/Http.svc/Translate"
)

var bingBaseURL *url.URL

type Bing struct {
	ServiceKey string
}

type bingResult struct {
	Translated string `xml:",chardata"`
}

func init() {
	var err error
	bingBaseURL, err = url.Parse(bingAPIBase)
	if err != nil {
		panic(err)
	}
}

func (b Bing) Translate(givenPhrase string, givenLang, targetLang language.Tag, out *chan Result) {
	requestURL := *bingBaseURL
	requestURL.RawQuery = "from=" + url.QueryEscape(givenLang.String()) + "&to=" + url.QueryEscape(targetLang.String()) + "&text=" + url.QueryEscape(givenPhrase)

	request, err := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	request.Header.Set("Ocp-Apim-Subscription-Key", b.ServiceKey)

	if err != nil {
		*out <- Result{
			Error: err,
		}
		close(*out)
		return
	}


	response, err := http.DefaultClient.Do(request)
	if err != nil {
		*out <- Result{
			Error: err,
		}
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	content := buf.String()

	if response.StatusCode != http.StatusOK {
		log.Printf("bing API returned error: %v", content)
		*out <- Result{
			Error: errors.New(response.Status),
		}
	}

	result := &bingResult{}
	err = xml.Unmarshal(buf.Bytes(), result)
	if err != nil {
		*out <- Result{
			Error: err,
		}
		return
	}

	*out <- Result{
		GivenLang:givenLang,
		GivenPhrase:givenPhrase,
		TargetLang:targetLang,
		TranslatedPhrase: result.Translated,
	}
}