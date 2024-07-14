package request

type ShortLinkRequest struct {
	Link      string `json:"link"`
	Expirable bool   `json:"expirable"`
}
