package request

type UpdateLinkRequest struct {
	Link      string `json:"link" valid:"required,url"`
	ShortLink string `json:"short_link" valid:"required,url"`
}
