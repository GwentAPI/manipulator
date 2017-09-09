package models

type GwentCard struct {
	Categories []string
	Faction    string
	Flavor     map[string]string
	Info       map[string]string
	IngameId   string
	Loyalties  []string
	Name       map[string]string
	Positions  []string
	Released   bool
	Strength   int
	Group      string `json:"type"`
	Variations map[string]GwentVariation
}

type GwentVariation struct {
	Art          GwentArt
	Availability string
	Collectible  bool
	Craft        GwentCost
	Mill         GwentCost
	Rarity       string
	VariationId  string
}

type GwentCost struct {
	Premium  int
	Standard int
}

type GwentArt struct {
	Artist    string
	High      string
	Low       string
	Medium    string
	Original  string
	Thumbnail string
}
