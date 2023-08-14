package helpers

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if url[:4] != "http" {
		return "http://" + url
	}
	return url
}

// To avoid unnecessary request spams to {localhost:3000}
func RemoveDomainError(url string) bool {
	if url == os.Getenv("API_DOMAIN") {
		return false
	}

	//removing common prefixes
	newUrl := strings.Replace(url, "http://", "", 1)
	newUrl = strings.Replace(newUrl, "https://", "", 1)
	newUrl = strings.Replace(newUrl, "www.", "", 1)
	newUrl = strings.Split(newUrl, "/")[0]
	if newUrl == os.Getenv("API_DOMAIN") {
		return false
	}

	return true

}
