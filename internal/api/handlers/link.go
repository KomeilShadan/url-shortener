package handlers

import (
	AppHttp "drto-link/internal/api/http"
	"drto-link/internal/api/request"
	"drto-link/internal/api/response"
	"drto-link/internal/service"
	"drto-link/internal/utils"
	"drto-link/pkg/log"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

	mongodb := ctx.MustGet("mongo").(*mongo.Client)

	_, err := mongodb.Database("link").Collection("links").InsertOne(ctx, bson.M{
		"link":       link,
		"short_link": shortLink,
	})
	if err != nil {
		log.Error(log.Mongodb, log.Insert, err, nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("database error"),
			Path:    ctx.FullPath(),
		})
	}
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
