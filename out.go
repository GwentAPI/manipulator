package main

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type GenericCollection struct {
	ID            bson.ObjectId "_id,omitempty"
	Name          string        `bson:"name"`
	UUID          []byte        `bson:"uuid"`
	Last_modified time.Time     `bson:"last_modified"`
}

type Card struct {
	ID            bson.ObjectId     "_id,omitempty"
	Categories    *[]string         "categories,omitempty"
	Faction       string            "faction"
	Flavor        map[string]string "flavor,omitempty"
	Info          map[string]string "info,omitempty"
	Strength      *int              "strength,omitempty"
	Name          map[string]string "name"
	Loyalties     []string          "loyalties,omitempty"
	Positions     []string          "positions"
	Faction_id    bson.ObjectId     "faction_id,omitempty"
	Group         string            "group"
	Group_id      bson.ObjectId     "group_id,omitempty"
	Categories_id []bson.ObjectId   "categories_id,omitempty"
	UUID          []byte            "uuid"
	Last_Modified time.Time         "last_modified"
}

type Variation struct {
	ID            bson.ObjectId "_id,omitempty"
	Card_id       bson.ObjectId "card_id,omitempty"
	Rarity_id     bson.ObjectId "rarity_id,omitempty"
	UUID          []byte
	Availability  string
	Rarity        string
	Craft         Cost      "craft,omitempty"
	Mill          Cost      "mill,omitempty"
	Art           Art       "art,omitempty"
	Last_Modified time.Time "last_modified"
}

type Cost struct {
	Normal  int
	Premium int
}

type Art struct {
	Artist         string  "artist,omitempty"
	FullsizeImage  *string "fullsizeImage"
	ThumbnailImage string  "thumbnailImage"
}
