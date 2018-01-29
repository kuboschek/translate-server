package main

import (
	"cloud.google.com/go/translate"
	"golang.org/x/net/context"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

type GcloudProvider struct {
	apiKey string
	client *translate.Client
}

func (p *GcloudProvider) makeGcloudClient() error {
	client, err := translate.NewClient(context.Background(), option.WithAPIKey(p.apiKey))
	if err != nil {
		return err
	}

	p.client = client
	return nil
}

func sendError(out *chan TranslateResult, err error) {
	*out <- TranslateResult{
		err: err,
	}
}

// GetTranslation translates the given text using the Google Cloud Translation API
func (p GcloudProvider) GetTranslation(givenPhrase, givenLang, targetLang string, out *chan TranslateResult) {
	if p.client == nil {
		err := p.makeGcloudClient()
		if err != nil {
			sendError(out, err)
			return
		}
	}

	targetTag, err := language.Parse(targetLang)
	if err != nil {
		sendError(out, err)
		return
	}

	sourceTag, err := language.Parse(givenLang)
	if err != nil {
		sendError(out, err)
		return
	}

	opts := translate.Options{
		Source: sourceTag,
		Format: translate.Text,
	}

	result, err := p.client.Translate(context.Background(), []string{givenPhrase}, targetTag, &opts)
	if err != nil {
		sendError(out, err)
		return
	}

	*out <- TranslateResult{
		givenLang:        givenLang,
		givenPhrase:      givenPhrase,
		targetLang:       targetLang,
		translatedPhrase: result[0].Text,
	}
}
