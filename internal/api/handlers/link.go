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
	"net/http"
)

func ShortLink(ctx *gin.Context) {
	cfg := config.Get()
	var (
		req              request.ShortLinkRequest
		shortLinkBaseURL string
		linkStruct       model.Link
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
	}
	link := utils.EnforceHTTP(req.Link)

	shortLinkBaseURL = cfg.App.ShortLinkBaseURL
	shortLink, _ := service.GenerateShortLink(link)
	updateQuery := bson.M{"$set": model.Link{Link: link, ShortLink: shortLinkBaseURL + shortLink}}

	mongodb := ctx.MustGet("mongo").(*mongo.Client)
	err := mongodb.Database("link").
		Collection("links").
		FindOneAndReplace(ctx, bson.M{"link": link}, updateQuery).
		Decode(&linkStruct)

	if err != nil {
		log.Error(log.Mongodb, log.Update, err, nil)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("database error"),
			Path:    ctx.FullPath(),
		})
	}
	//implement redis store or caching

	ctx.JSON(http.StatusCreated, AppHttp.ApiResponse{
		Data: response.ShortLinkResponse{
			Link:      linkStruct.Link,
			ShortLink: linkStruct.ShortLink,
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
	}
	//implement fetching link from mongo or redis

	ctx.JSON(http.StatusOK, AppHttp.ApiResponse{
		Data: response.ResolveLinkResponse{
			Link: linkStruct.Link,
		},
	})
}
