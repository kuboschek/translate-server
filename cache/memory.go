package cache

import (
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

type memoryCache map[string]map[language.Tag]string

var (
	Memory         memoryCache
	phraseNotFound   = errors.New("phrase not found in memory cache")
)

func init() {
	Memory = make(memoryCache)
}

func (p memoryCache) Put(sourcePhrase string, targetLang language.Tag, targetPhrase string) error {

	if p[sourcePhrase] == nil {
		p[sourcePhrase] = make(map[language.Tag]string)
	}

	p[sourcePhrase][targetLang] = targetPhrase

	return nil
}

func (p memoryCache) Has(sourcePhrase string, targetLang language.Tag) bool {
	phrases, hasPhrase := p[sourcePhrase]
	_, hasTargetLang := phrases[targetLang]
	return hasPhrase && hasTargetLang
}

func (p memoryCache) Get(sourcePhrase string, targetLang language.Tag) (targetPhrase string, err error) {
	phrases, ok := p[sourcePhrase]
	if !ok {
		return "", phraseNotFound
	}

	phrase, ok := phrases[targetLang]
	if !ok {
		return "", phraseNotFound
	}

	return phrase, nil
}
