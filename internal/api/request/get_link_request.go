package request

type GetLinkRequest struct {
	ShortLink string `json:"short_link" valid:"required,url"`
}
