package upstream

import (
	"cloud.google.com/go/translate"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// GoogleProvider is a Service that
type GoogleProvider struct {
	Key    string
	client *translate.Client
}

// makeGoogleClient sets up the Google Translation Client library
func (p *GoogleProvider) makeGoogleClient() error {
	client, err := translate.NewClient(context.Background(), option.WithAPIKey(p.Key))
	if err != nil {
		return err
	}

	p.client = client
	return nil
}

func sendError(out *chan Result, err error) {
	*out <- Result{
		Error: err,
	}
}

// Translate translates the given text using the Google Cloud Translation API
// The translation result is sent to channel out.
func (p GoogleProvider) Translate(givenPhrase string, givenLang, targetLang language.Tag, out *chan Result) {
	if p.client == nil {
		err := p.makeGoogleClient()
		if err != nil {
			sendError(out, err)
			return
		}
	}

	opts := translate.Options{
		Source: givenLang,
		Format: translate.Text,
	}

	result, err := p.client.Translate(context.Background(), []string{givenPhrase}, targetLang, &opts)
	if err != nil {
		sendError(out, err)
		return
	}

	*out <- Result{
		GivenLang:        givenLang,
		GivenPhrase:      givenPhrase,
		TargetLang:       targetLang,
		TranslatedPhrase: result[0].Text,
	}
}
