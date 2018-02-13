package upstream

import (
	"golang.org/x/text/language"
	"net/url"
	"net/http"
	"bytes"
)

const (
	bingAPIBase = "https://api.microsofttranslator.com/v2/Http.svc/Translate"
)

var bingBaseURL *url.URL

type Bing struct {
	ServiceKey string
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
	requestURL.RawQuery = "from=" + givenLang.String() + "&to=" + targetLang.String() + "&text=" + givenPhrase

	response, err := http.Get(requestURL.String())
	if err != nil {
		*out <- Result{
			Error: err,
		}
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	content := buf.String()

	*out <- Result{
		GivenLang:givenLang,
		GivenPhrase:givenPhrase,
		TargetLang:targetLang,
		TranslatedPhrase: content,
	}
}