package shortener_test

import (
	"testing"

	"github.com/likhi/url-shortener/pkg/shortener"
)

func TestGenerate_Length(t *testing.T) {
	for _, length := range []int{5, 7, 10, 20} {
		code, err := shortener.Generate(length)
		if err != nil {
			t.Fatalf("Generate(%d) error: %v", length, err)
		}
		if len(code) != length {
			t.Errorf("Generate(%d) = %q, want len %d", length, code, length)
		}
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for range 1000 {
		code, err := shortener.Generate(7)
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q generated", code)
		}
		seen[code] = true
	}
}

func TestGenerate_Charset(t *testing.T) {
	for range 500 {
		code, err := shortener.Generate(7)
		if err != nil {
			t.Fatalf("Generate error: %v", err)
		}
		for _, ch := range code {
			if !isAlphanumeric(ch) {
				t.Fatalf("Generate produced invalid char %q in %q", ch, code)
			}
		}
	}
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
