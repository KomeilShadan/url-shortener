package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
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
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}

func AvoidDSelfDomain(url string) bool {
	// Basically this function removes all the commonly found
	// prefixes from URL such as http, https, www
	// then checks of the remaining string is the APP_HOST itself
	if url == os.Getenv("APP_HOST") {
		return false
	}
	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	if newURL == os.Getenv("APP_HOST") {
		return false
	}
	return true
}
