package handlers

import (
	AppHttp "drto-link/internal/api/http"
	"drto-link/internal/api/request"
	"drto-link/internal/api/response"
	"drto-link/internal/service"
	"drto-link/internal/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ShortLink(ctx *gin.Context) {
	var req request.ShortLinkRequest

	// Bind the JSON payload to the ShortLinkRequest struct
	utils.BindRequestBody(ctx, &req)

	// Validate the ShortLinkRequest struct using Govalidator
	utils.ValidateRequestBody(ctx, &req)

	if !utils.AvoidDSelfDomain(req.Link) {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, AppHttp.ApiResponse{
			Message: "Unprocessable Entity (Nice Try!)",
			Error:   errors.New("unprocessable input link"),
			Path:    ctx.FullPath(),
		})
	}
	link := utils.EnforceHTTP(req.Link)

	shortLink, _ := service.GenerateShortLink(link)

	//implement mongodb store
	//implement redis store or caching

	ctx.JSON(http.StatusCreated, AppHttp.ApiResponse{
		Data: response.ShortLinkResponse{
			Link:      req.Link,
			ShortLink: shortLink,
			Expirable: req.Expirable,
		},
	})
}

func ResolveLink(ctx *gin.Context) {
	var req request.ResolveLinkRequest

	utils.ValidateRequestBody(ctx, &req)

	utils.ValidateRequestBody(ctx, &req)

	//implement fetching link from mongo or redis
	link := ""

	ctx.JSON(http.StatusOK, AppHttp.ApiResponse{
		Data: response.ResolveLinkResponse{
			Link: link,
		},
	})
}
