package service

import (
	"strings"
	"testing"
)

func TestGenerateShortLinkHash_Success(t *testing.T) {
	tests := []struct {
		name string
		link string
	}{
		{
			name: "Simple URL",
			link: "https://example.com",
		},
		{
			name: "URL with path",
			link: "https://example.com/path/to/resource",
		},
		{
			name: "URL with query params",
			link: "https://example.com?param1=value1&param2=value2",
		},
		{
			name: "Long URL",
			link: "https://example.com/very/long/path/with/many/segments/and/parameters?param1=value1&param2=value2&param3=value3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GenerateShortLinkHash(tt.link)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if hash == "" {
				t.Error("Expected non-empty hash")
			}

			if len(hash) != 7 {
				t.Errorf("Expected hash length 7, got %d", len(hash))
			}
		})
	}
}

func TestGenerateShortLinkHash_EmptyLink(t *testing.T) {
	hash, err := GenerateShortLinkHash("")

	if err == nil {
		t.Error("Expected error for empty link")
	}

	if hash != "" {
		t.Errorf("Expected empty hash, got: %s", hash)
	}

	if err.Error() != "empty link" {
		t.Errorf("Expected 'empty link' error, got: %v", err)
	}
}

func TestGenerateShortLinkHash_Consistency(t *testing.T) {
	link := "https://example.com/test"

	hash1, err1 := GenerateShortLinkHash(link)
	hash2, err2 := GenerateShortLinkHash(link)

	if err1 != nil || err2 != nil {
		t.Fatalf("Expected no errors, got: %v, %v", err1, err2)
	}

	if hash1 != hash2 {
		t.Errorf("Expected consistent hashes for same link. Got %s and %s", hash1, hash2)
	}
}

func TestGenerateShortLinkHash_Uniqueness(t *testing.T) {
	links := []string{
		"https://example.com/page1",
		"https://example.com/page2",
		"https://example.com/page3",
		"https://different.com/page1",
	}

	hashes := make(map[string]string)

	for _, link := range links {
		hash, err := GenerateShortLinkHash(link)
		if err != nil {
			t.Fatalf("Failed to generate hash for %s: %v", link, err)
		}

		if existingLink, exists := hashes[hash]; exists {
			t.Errorf("Hash collision detected: %s and %s both generated hash %s", link, existingLink, hash)
		}

		hashes[hash] = link
	}
}

func TestGenerateShortLinkHash_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name string
		link string
	}{
		{
			name: "Leading whitespace",
			link: "  https://example.com",
		},
		{
			name: "Trailing whitespace",
			link: "https://example.com  ",
		},
		{
			name: "Both whitespace",
			link: "  https://example.com  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GenerateShortLinkHash(tt.link)

			// Should handle whitespace gracefully
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if hash == "" {
				t.Error("Expected non-empty hash")
			}
		})
	}
}

func TestGenerateShortLinkHash_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name string
		link string
	}{
		{
			name: "URL with special chars",
			link: "https://example.com/path?query=hello%20world&foo=bar#section",
		},
		{
			name: "URL with unicode",
			link: "https://example.com/путь/к/ресурсу",
		},
		{
			name: "URL with emoji",
			link: "https://example.com/😀/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GenerateShortLinkHash(tt.link)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if hash == "" {
				t.Error("Expected non-empty hash")
			}

			if len(hash) != 7 {
				t.Errorf("Expected hash length 7, got %d", len(hash))
			}
		})
	}
}

func TestGenerateShortLinkHash_OnlyWhitespace(t *testing.T) {
	hash, err := GenerateShortLinkHash("   ")

	// Depending on implementation, this might be treated as empty
	// The current implementation checks EmptyString which might trim
	if err == nil {
		t.Error("Expected error for whitespace-only link")
	}

	if hash != "" {
		t.Errorf("Expected empty hash, got: %s", hash)
	}
}

func BenchmarkGenerateShortLinkHash(b *testing.B) {
	link := "https://example.com/test/path/to/resource"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateShortLinkHash(link)
	}
}

func BenchmarkGenerateShortLinkHash_LongURL(b *testing.B) {
	link := "https://example.com/" + strings.Repeat("very/long/path/", 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateShortLinkHash(link)
	}
}
