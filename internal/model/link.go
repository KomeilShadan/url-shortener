package model

type Link struct {
	OriginalURL   string `json:"link" bson:"link"`
	ShortLinkHash string `json:"short_link" bson:"short_link_hash"`
}
