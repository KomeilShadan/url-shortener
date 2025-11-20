package service

import (
	"errors"
	"janus/internal/utils"
)

// GenerateShortLinkHash creates a 7-character hash from a URL using SHA-256 and Base64 encoding.
//
// The hash generation process:
//  1. Validates the input URL is not empty
//  2. Computes SHA-256 hash of the URL
//  3. Encodes the hash using Base64
//  4. Returns the first 7 characters as the short link identifier
//
// The same URL will always produce the same hash (deterministic).
// Returns an error if the URL is empty or hashing fails.
func GenerateShortLinkHash(link string) (string, error) {
	if utils.EmptyString(link) {
		return "", errors.New("empty link")
	}

	linkHash, err := utils.Sha256Of(link)
	if err != nil {
		return "", err
	}

	shortLink := utils.Base64Encode(linkHash)

	return shortLink[:7], nil
}
