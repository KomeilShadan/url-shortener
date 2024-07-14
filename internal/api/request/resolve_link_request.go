package request

type ResolveLinkRequest struct {
	ShortLink string `json:"short_link" valid:"required,url"`
}
