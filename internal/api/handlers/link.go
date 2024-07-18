package handlers

import (
	AppHttp "drto-link/internal/api/http"
	"drto-link/internal/api/request"
	"drto-link/internal/api/response"
	"drto-link/internal/config"
	"drto-link/internal/model"
	"drto-link/internal/service"
	"drto-link/internal/utils"
	"drto-link/pkg/log"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
)

func ShortLink(ctx *gin.Context) {
	cfg := config.Get()
	var (
		req              request.ShortLinkRequest
		shortLinkBaseURL string
		fullShortLink    string
	)

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
		return
	}
	link := utils.EnforceHTTP(req.Link)

	shortLinkBaseURL = cfg.App.ShortLinkBaseURL
	shortLink, _ := service.GenerateShortLink(link)
	fullShortLink = shortLinkBaseURL + shortLink
	updateQuery := bson.M{"link": link, "short_link": fullShortLink}

	mongodb := ctx.MustGet("mongo").(*mongo.Client)
	opts := options.Replace().SetUpsert(true)

	_, err := mongodb.Database("link").
		Collection("links").
		ReplaceOne(ctx, bson.M{"link": link}, updateQuery, opts)

	if err != nil {
		log.Error(log.Mongodb, log.Update, err, nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("database error"),
			Path:    ctx.FullPath(),
		})
		return
	}
	//implement redis store or caching

	ctx.JSON(http.StatusCreated, AppHttp.ApiResponse{
		Data: response.ShortLinkResponse{
			Link:      link,
			ShortLink: fullShortLink,
			Expirable: req.Expirable,
		},
	})
}

func ResolveLink(ctx *gin.Context) {
	var (
		req        request.ResolveLinkRequest
		linkStruct model.Link
	)

	utils.BindRequestBody(ctx, &req)

	utils.ValidateRequestBody(ctx, &req)

	mongodb := ctx.MustGet("mongo").(*mongo.Client)
	err := mongodb.Database("link").
		Collection("links").
		FindOne(ctx, bson.M{"short_link": req.ShortLink}).
		Decode(&linkStruct)

	if err != nil {
		log.Error(log.Mongodb, log.Select, err, nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("database error"),
			Path:    ctx.FullPath(),
		})
		return
	}
	//implement fetching link from redis

	ctx.JSON(http.StatusOK, AppHttp.ApiResponse{
		Data: response.ResolveLinkResponse{
			Link: linkStruct.Link,
		},
	})
}
