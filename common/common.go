package common

import (
	"github.com/rainycape/unidecode"
	"regexp"
	"strings"
)

func GetArtUrl(cardName string) string {
	var re = regexp.MustCompile("[^a-z0-9]+")
	cardName = unidecode.Unidecode(cardName)
	return strings.Trim(re.ReplaceAllString(strings.ToLower(cardName), "-"), "-")
}

/*
type DBInterface interface {
	InsertGenericCollection(collectionName string, names map[string]struct{})
	EnsureSimpleUniqueIndex(key string, name string) error
	EnsureSimpleIndex(key string, name string, isUnique bool) error
	InsertCard(collectionName string, cards map[string]models.GwentCard)
	InsertVariation(collectionName string, cards map[string]models.GwentCard)
}
*/
