package utils

import (
	"crypto/sha256"
	AppHttp "drto-link/internal/api/http"
	"drto-link/internal/config"
	"encoding/base64"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func EmptyString(str string) bool {
	str = strings.TrimSpace(str)
	return strings.EqualFold(str, "")
}

func Sha256Of(input string) ([]byte, error) {
	algorithm := sha256.New()
	_, err := algorithm.Write([]byte(strings.TrimSpace(input)))
	if err != nil {
		return nil, err
	}
	return algorithm.Sum(nil), nil
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func EnforceHTTP(url string) string {
	url = strings.TrimSpace(url)

	if url[:4] != "http" {
		urlParts := strings.Split(url, "://")
		urlWithoutScheme := urlParts[len(urlParts)-1]
		return "http://" + strings.Trim(urlWithoutScheme, "/")
	}
	return url
}

func AvoidDSelfDomain(url string) bool {
	// Basically this function removes all the commonly found
	// prefixes from URL such as http, https, www
	// then checks of the remaining string is the APP_HOST itself
	var host string = config.Get().App.Host
	if url == host {
		return false
	}
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	if newURL == host {
		return false
	}
	return true
}

func BindRequestBody(ctx *gin.Context, req interface{}) {
	err := ctx.BindJSON(req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, AppHttp.ApiResponse{
			Message: "Bad Request",
			Error:   errors.New("malformed request body").Error(),
			Path:    ctx.FullPath(),
		})
		return
	}
}

func ValidateRequestBody(ctx *gin.Context, req interface{}) {
	_, err := govalidator.ValidateStruct(req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, AppHttp.ApiResponse{
			Message: "Bad Request",
			Error:   errors.New("invalid request body").Error(),
			Path:    ctx.FullPath(),
		})
		return
	}
}
