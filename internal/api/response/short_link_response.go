package response

type ShortLinkResponse struct {
	Link      string `json:"link"`
	ShortLink string `json:"short_link"`
	Expirable bool   `json:"expirable"`
	ExpiresAt uint32 `json:"expires_at"`
}
