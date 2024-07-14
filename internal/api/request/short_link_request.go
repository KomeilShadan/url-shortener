package request

type ShortLinkRequest struct {
	Link      string `json:"link" valid:"required,url"`
	Expirable bool   `json:"expirable" valid:"optional,bool"`
}
