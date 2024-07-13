package service

import (
	"drto-link/internal/utils"
	"errors"
)

func GenerateShortLink(link string) (string, error) {
	var (
		linkHash  []byte
		shortLink string
		err       error
	)
	if utils.EmptyString(link) {
		return "", errors.New("empty link")
	}
	linkHash, err = utils.Sha256Of(link)
	if err != nil {
		return "", err
	}
	shortLink = utils.Base64Encode(linkHash)
	return shortLink[:7], nil
}
