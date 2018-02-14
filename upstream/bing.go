package upstream

import (
	"bytes"
	"encoding/xml"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"net/url"
)

const (
	azureAPIBase = "https://api.microsofttranslator.com/v2/Http.svc/Translate"
)

var azureBaseURL *url.URL

// Azure represents a translation service calling the Azure Cognitive Services Machine Translation Service.
type Azure struct {
	// ServiceKey is the key from the Azure dashboard
	ServiceKey string
}

type bingResult struct {
	// Translated is the translation content as returned from Azure
	Translated string `xml:",chardata"`
}

func init() {
	var err error
	azureBaseURL, err = url.Parse(azureAPIBase)
	if err != nil {
		panic(err)
	}
}

// Translate call Microsoft Cognitive Services to translate the given string
func (b Azure) Translate(givenPhrase string, givenLang, targetLang language.Tag, out *chan Result) {
	requestURL := *azureBaseURL
	requestURL.RawQuery = "from=" + url.QueryEscape(givenLang.String()) + "&to=" + url.QueryEscape(targetLang.String()) + "&text=" + url.QueryEscape(givenPhrase)

	request, err := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	request.Header.Set("Ocp-Apim-Subscription-Key", b.ServiceKey)

	if err != nil {
		*out <- Result{
			Error: err,
		}
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
		log.Printf("Azure API returned error: %v", content)
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
		GivenLang:        givenLang,
		GivenPhrase:      givenPhrase,
		TargetLang:       targetLang,
		TranslatedPhrase: result.Translated,
	}
}
