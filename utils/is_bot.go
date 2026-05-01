package utils

import (
	"strings"
)

func IsBot(userAgent string) bool {
	bots := []string{
		"facebookexternalhit",
		"Facebot",
		"Twitterbot",
		"LinkedInBot",
		"TelegramBot",
		"WhatsApp",
	}

	for _, bot := range bots {
		if strings.Contains(strings.ToLower(userAgent), strings.ToLower(bot)) {
			return true
		}
	}
	return false
}
