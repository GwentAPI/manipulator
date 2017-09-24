package database

import (
	"crypto/tls"
	"github.com/GwentAPI/gwentapi/manipulator/common"
	"github.com/GwentAPI/gwentapi/manipulator/models"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"strconv"
	"time"
)

const DOMAIN string = "46bf3452-28e7-482c-9bbf-df053873b021"

type ReposClient struct{}

type MongoConnectionSettings struct {
	Host                   []string
	Db                     string
	AuthenticationDatabase string
	Username               string
	Password               string
	UseSSL                 bool
	Timeout                time.Duration
}

func (c ReposClient) CreateSession(authInfo MongoConnectionSettings) (*mgo.Session, error) {

	tlsConfig := &tls.Config{}

	dialInfo := &mgo.DialInfo{
		Addrs:    authInfo.Host,
		Database: authInfo.Db,
		Source:   authInfo.AuthenticationDatabase,
		Username: authInfo.Username,
		Password: authInfo.Password,
		Timeout:  authInfo.Timeout,
	}

	if authInfo.UseSSL {
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

func (c ReposClient) InsertGenericCollection(db *mgo.Database, collectionName string, names map[string]struct{}) {
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
		generic := models.GenericCollection{}

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

func (c ReposClient) EnsureSimpleUniqueIndex(collection *mgo.Collection, key string, name string) error {
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

func (c ReposClient) EnsureSimpleIndex(collection *mgo.Collection, key string, name string, isUnique bool) error {
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

func (c ReposClient) InsertCard(db *mgo.Database, collectionName string, cards map[string]models.GwentCard) {
	domainUUID, err := uuid.FromString(DOMAIN)
	if err != nil {
		log.Fatal("DomainUUID error: ", err)
	}

	collection := db.C(collectionName)

	errNameIndex := c.EnsureSimpleUniqueIndex(collection, "name.en-US", "name.en-US")
	errUuidIndex := c.EnsureSimpleUniqueIndex(collection, "uuid", "uuid")
	if errNameIndex != nil || errUuidIndex != nil {
		log.Fatal("Error creating index: ", errNameIndex, " ", errUuidIndex)
	}
	c.EnsureSimpleUniqueIndex(collection, "name.de-DE", "name.de-DE")
	c.EnsureSimpleUniqueIndex(collection, "name.fr-FR", "name.fr-FR")
	c.EnsureSimpleUniqueIndex(collection, "name.pl-PL", "name.pl-PL")
	c.EnsureSimpleUniqueIndex(collection, "name.pt-BR", "name.pt-BR")
	// Special snowflakes that actually have duplicates names for the same localization.
	c.EnsureSimpleIndex(collection, "name.zh-TW", "name.zh-TW", false)
	c.EnsureSimpleUniqueIndex(collection, "name.es-ES", "name.es-ES")
	c.EnsureSimpleUniqueIndex(collection, "name.es-MX", "name.es-MX")
	c.EnsureSimpleUniqueIndex(collection, "name.it-IT", "name.it-IT")
	c.EnsureSimpleUniqueIndex(collection, "name.ja-JP", "name.ja-JP")
	c.EnsureSimpleUniqueIndex(collection, "name.ru-RU", "name.ru-RU")
	c.EnsureSimpleUniqueIndex(collection, "name.zh-CN", "name.zh-CN")

	bulk := collection.Bulk()
	bulk.Unordered()

	factionIDs := map[string]bson.ObjectId{}
	groupIDs := map[string]bson.ObjectId{}
	categoryIDs := map[string]bson.ObjectId{}

	for _, v := range cards {
		c := models.Card{
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
			queryResult := models.GenericCollection{}
			db.C("factions").Find(bson.M{"name": v.Faction}).Select(bson.M{"_id": 1}).One(&queryResult)
			factionIDs[v.Faction] = queryResult.ID
			c.Faction_id = queryResult.ID
		}

		if groupID, ok := groupIDs[v.Group]; ok {
			c.Group_id = groupID
		} else {
			queryResult := models.GenericCollection{}
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
					queryResult := models.GenericCollection{}
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

func (c ReposClient) InsertVariation(db *mgo.Database, collectionName string, cards map[string]models.GwentCard) {
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
		queryResult := models.Card{}
		db.C("cards").Find(bson.M{"name.en-US": card.Name["en-US"]}).Select(bson.M{"_id": 1}).One(&queryResult)
		artUrl := common.GetArtUrl(card.Name["en-US"])
		//log.Println(artUrl)
		var numVariation int = 0
		for _, variation := range card.Variations {
			// UUID : name + availability
			numVariation++
			thumbnailUrl := artUrl + "-" + strconv.Itoa(numVariation) + "-thumbnail.png"
			mediumSizeUrl := artUrl + "-" + strconv.Itoa(numVariation) + "-medium.png"
			originalSizeUrl := artUrl + "-" + strconv.Itoa(numVariation) + "-full.png"

			v := models.Variation{
				UUID:         uuid.NewV5(domainUUID, card.Name["en-US"]+variation.Availability).Bytes(),
				Availability: variation.Availability,
				Rarity:       variation.Rarity,
				Craft: models.Cost{
					Normal:  variation.Craft.Standard,
					Premium: variation.Craft.Premium,
				},
				Mill: models.Cost{
					Normal:  variation.Mill.Standard,
					Premium: variation.Mill.Premium,
				},
				Art: models.Art{
					FullsizeImage:   &originalSizeUrl,
					MediumsizeImage: mediumSizeUrl,
					ThumbnailImage:  thumbnailUrl,
					Artist:          variation.Art.Artist,
				},
				Last_Modified: time.Now().UTC(),
			}
			v.Card_id = queryResult.ID

			queryRarityResult := models.GenericCollection{}
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
