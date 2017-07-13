package main

import (
	"crypto/tls"
	"github.com/rainycape/unidecode"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

type Authentication struct {
	Source   string
	Username string
	Password string
}

func CreateSession(addrs []string, database string, authInfo Authentication, useSSl bool, timeout time.Duration) (*mgo.Session, error) {

	tlsConfig := &tls.Config{}

	dialInfo := &mgo.DialInfo{
		Addrs:    addrs,
		Database: database,
		Source:   authInfo.Source,
		Username: authInfo.Username,
		Password: authInfo.Password,
		Timeout:  timeout,
	}

	if useSSl {
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	session.SetSocketTimeout(10 * time.Second)
	session.SetMode(mgo.Strong, true)
	session.SetSafe(&mgo.Safe{WMode: "majority"})
	return session, nil
}

func InsertGenericCollection(db *mgo.Database, collectionName string, names map[string]struct{}) {
	domainUUID, err := uuid.FromString(DOMAIN)
	if err != nil {
		log.Fatal("DomainUUID error: ", err)
	}

	collection := db.C(collectionName)

	nameIndex := mgo.Index{
		Key:        []string{"name"},
		Unique:     true,
		Background: true,
		Name:       "name",
	}

	uuidIndex := mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		Background: true,
		Name:       "uuid",
	}

	errNameIndex := collection.EnsureIndex(nameIndex)
	errUuidIndex := collection.EnsureIndex(uuidIndex)
	if errNameIndex != nil || errUuidIndex != nil {
		log.Fatal("Error creating index: ", errNameIndex, " ", errUuidIndex)
	}

	bulk := collection.Bulk()
	bulk.Unordered()

	for key, _ := range names {
		generic := GenericCollection{}

		generic.Name = key
		generic.Last_modified = time.Now().UTC()
		generic.UUID = uuid.NewV5(domainUUID, key).Bytes()

		selector := bson.M{"uuid": generic.UUID}
		bulk.Upsert(selector, generic)
	}

	_, bulkErr := bulk.Run()
	if bulkErr != nil {
		log.Fatal("Error bulk upsert: ", bulkErr)
	}
}

func EnsureSimpleUniqueIndex(collection *mgo.Collection, key string, name string) error {
	index := mgo.Index{
		Key:        []string{key},
		Unique:     true,
		Background: true,
		Name:       name,
	}
	err := collection.EnsureIndex(index)
	if err != nil {
		log.Println("Problem with index key ", key, " with name ", name, " : ", err)
	}
	return err
}

func EnsureSimpleIndex(collection *mgo.Collection, key string, name string, isUnique bool) error {
	index := mgo.Index{
		Key:        []string{key},
		Unique:     isUnique,
		Background: true,
		Name:       name,
	}
	err := collection.EnsureIndex(index)
	if err != nil {
		log.Println("Problem with index key ", key, " with name ", name, " : ", err)
	}
	return err
}

func InsertCard(db *mgo.Database, collectionName string, cards map[string]GwentCard) {
	domainUUID, err := uuid.FromString(DOMAIN)
	if err != nil {
		log.Fatal("DomainUUID error: ", err)
	}

	collection := db.C(collectionName)

	errNameIndex := EnsureSimpleUniqueIndex(collection, "name.en-US", "name.en-US")
	errUuidIndex := EnsureSimpleUniqueIndex(collection, "uuid", "uuid")
	if errNameIndex != nil || errUuidIndex != nil {
		log.Fatal("Error creating index: ", errNameIndex, " ", errUuidIndex)
	}
	EnsureSimpleUniqueIndex(collection, "name.de-DE", "name.de-DE")
	EnsureSimpleUniqueIndex(collection, "name.fr-FR", "name.fr-FR")
	EnsureSimpleUniqueIndex(collection, "name.pl-PL", "name.pl-PL")
	EnsureSimpleUniqueIndex(collection, "name.pt-BR", "name.pt-BR")
	// Special snowflakes that actually have duplicates names for the same localization.
	EnsureSimpleIndex(collection, "name.zh-TW", "name.zh-TW", false)
	EnsureSimpleUniqueIndex(collection, "name.es-ES", "name.es-ES")
	EnsureSimpleUniqueIndex(collection, "name.es-MX", "name.es-MX")
	EnsureSimpleUniqueIndex(collection, "name.it-IT", "name.it-IT")
	EnsureSimpleUniqueIndex(collection, "name.ja-JP", "name.ja-JP")
	EnsureSimpleUniqueIndex(collection, "name.ru-RU", "name.ru-RU")
	EnsureSimpleUniqueIndex(collection, "name.zh-CN", "name.zh-CN")

	bulk := collection.Bulk()
	bulk.Unordered()

	factionIDs := map[string]bson.ObjectId{}
	groupIDs := map[string]bson.ObjectId{}
	categoryIDs := map[string]bson.ObjectId{}

	for _, v := range cards {
		c := Card{
			Name:          v.Name,
			UUID:          uuid.NewV5(domainUUID, v.Name["en-US"]).Bytes(),
			Group:         v.Group,
			Faction:       v.Faction,
			Positions:     v.Positions,
			Last_Modified: time.Now().UTC(),
		}

		if v.Strength > 0 {
			c.Strength = new(int)
			*c.Strength = v.Strength
		}

		if _, ok := v.Info["en-US"]; ok {
			c.Info = v.Info
		}
		if _, ok := v.Flavor["en-US"]; ok {
			c.Flavor = v.Flavor
		}

		if len(v.Loyalties) > 0 {
			c.Loyalties = v.Loyalties
		}

		if factionID, ok := factionIDs[v.Faction]; ok {
			c.Faction_id = factionID
		} else {
			queryResult := GenericCollection{}
			db.C("factions").Find(bson.M{"name": v.Faction}).Select(bson.M{"_id": 1}).One(&queryResult)
			factionIDs[v.Faction] = queryResult.ID
			c.Faction_id = queryResult.ID
		}

		if groupID, ok := groupIDs[v.Group]; ok {
			c.Group_id = groupID
		} else {
			queryResult := GenericCollection{}
			db.C("groups").Find(bson.M{"name": v.Group}).Select(bson.M{"_id": 1}).One(&queryResult)
			groupIDs[v.Group] = queryResult.ID
			c.Group_id = queryResult.ID
		}

		if len(v.Categories) > 0 {
			c.Categories = new([]string)
			*c.Categories = v.Categories

			for _, category := range v.Categories {
				if categoryID, ok := categoryIDs[category]; ok {
					c.Categories_id = append(c.Categories_id, categoryID)
				} else {
					queryResult := GenericCollection{}
					db.C("categories").Find(bson.M{"name": category}).Select(bson.M{"_id": 1}).One(&queryResult)
					categoryIDs[category] = queryResult.ID
					c.Categories_id = append(c.Categories_id, queryResult.ID)
				}
			}
		} else {

		}
		selector := bson.M{"uuid": c.UUID}
		bulk.Upsert(selector, c)
	}

	_, bulkErr := bulk.Run()
	if bulkErr != nil {
		log.Fatal("Error bulk card upsert: ", bulkErr)
	}
}

func InsertVariation(db *mgo.Database, collectionName string, cards map[string]GwentCard) {
	domainUUID, err := uuid.FromString(DOMAIN)
	if err != nil {
		log.Fatal("DomainUUID error: ", err)
	}

	collection := db.C(collectionName)

	cardIndex := mgo.Index{
		Key:        []string{"card_id"},
		Unique:     true,
		Background: true,
		Name:       "card_id",
	}

	uuidIndex := mgo.Index{
		Key:        []string{"uuid"},
		Unique:     true,
		Background: true,
		Name:       "uuid",
	}

	errNameIndex := collection.EnsureIndex(cardIndex)
	errUuidIndex := collection.EnsureIndex(uuidIndex)
	if errNameIndex != nil || errUuidIndex != nil {
		log.Fatal("Error creating index: ", errNameIndex, " ", errUuidIndex)
	}

	bulk := collection.Bulk()
	bulk.Unordered()

	for _, card := range cards {
		queryResult := Card{}
		db.C("cards").Find(bson.M{"name.en-US": card.Name["en-US"]}).Select(bson.M{"_id": 1}).One(&queryResult)
		artUrl := GetArtUrl(card.Name["en-US"])
		//log.Println(artUrl)
		thumbnailUrl := artUrl + "-thumbnail.png"
		originalSizeUrl := artUrl + "-full.png"

		for _, variation := range card.Variations {
			// UUID : name + availability
			v := Variation{
				UUID:         uuid.NewV5(domainUUID, card.Name["en-US"]+variation.Availability).Bytes(),
				Availability: variation.Availability,
				Rarity:       variation.Rarity,
				Craft: Cost{
					Normal:  variation.Craft.Standard,
					Premium: variation.Craft.Premium,
				},
				Mill: Cost{
					Normal:  variation.Mill.Standard,
					Premium: variation.Mill.Premium,
				},
				Art: Art{
					FullsizeImage:  &originalSizeUrl,
					ThumbnailImage: thumbnailUrl,
					Artist:         variation.Art.Artist,
				},
				Last_Modified: time.Now().UTC(),
			}
			v.Card_id = queryResult.ID

			queryRarityResult := GenericCollection{}
			db.C("rarities").Find(bson.M{"name": v.Rarity}).Select(bson.M{"_id": 1}).One(&queryRarityResult)

			v.Rarity_id = queryRarityResult.ID

			selector := bson.M{"uuid": v.UUID}
			bulk.Upsert(selector, v)
		}
	}
	_, bulkErr := bulk.Run()
	if bulkErr != nil {
		log.Fatal("Error bulk variation upsert: ", bulkErr)
	}
}

func GetArtUrl(cardName string) string {
	var re = regexp.MustCompile("[^a-z0-9]+")
	cardName = unidecode.Unidecode(cardName)
	return strings.Trim(re.ReplaceAllString(strings.ToLower(cardName), "-"), "-")
}
