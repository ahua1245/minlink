package util

import (
	"net/url"
	"regexp"
)

func IsValidURL(str string) bool {
	if len(str) == 0 || len(str) > 2048 {
		return false
	}

	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)
	if !hostnameRegex.MatchString(u.Hostname()) {
		return false
	}

	return true
}

func IsValidShortCode(str string) bool {
	if len(str) < 6 || len(str) > 12 {
		return false
	}
	regex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return regex.MatchString(str)
}