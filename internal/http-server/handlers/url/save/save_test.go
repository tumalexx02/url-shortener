package save_test

import (
	"github.com/go-playground/validator/v10"
	"testing"
	"url-shortner/internal/http-server/handlers/url/save"
)

func TestSave_FormatUrl(t *testing.T) {
	validUrls := []string{
		"http://www.example.com",
		"https://www.example.com",
		"example.com",
		"www.example.com",
		"https://example.com/path",
		"http://sub.example.com",
	}

	notValidUrls := []string{
		"http://www.example",
		"example",
		"http://.com",
		"ftp://example.com", // only http/https
		"127.0.0.1",
		"http://localhost",
	}

	val := validator.New()
	_ = val.RegisterValidation("custom_url", save.ValidateURL)

	for _, url := range validUrls {
		t.Run("valid_"+url, func(t *testing.T) {
			if err := val.Var(url, "custom_url"); err != nil {
				t.Errorf("not valid url = %s", url)
			}
		})
	}

	for _, url := range notValidUrls {
		t.Run("not_valid_"+url, func(t *testing.T) {
			if err := val.Var(url, "custom_url"); err == nil {
				t.Errorf("valid url = %s", url)
			}
		})
	}
}
