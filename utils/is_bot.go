package utils

import (
	"net"
	"strings"
)

var compiledNetworks []*net.IPNet

func init() {
	rawNetworks := []string{
		// Rangos masivos de Meta (Facebook/Instagram)
		"66.220.144.0/20", // Cubre la 66.220.149.x que se filtró
		"173.252.64.0/18",
		"31.13.64.0/18",
		"57.141.0.0/16",
		// Rangos masivos de AWS (Donde viven los bots de preview)
		"52.0.0.0/8", // Cubre las 52.38.x y 52.42.x que se filtraron
		"16.0.0.0/8", // Cubre la 16.144.x que se filtró
		"34.0.0.0/8",
		"35.0.0.0/8",
		"54.0.0.0/8",
		// LinkedIn & Twitter
		"108.174.0.0/20",
		"199.16.156.0/22",
	}

	for _, cidr := range rawNetworks {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			compiledNetworks = append(compiledNetworks, network)
		}
	}
}

func IsBot(ua string, ipStr string) bool {
	lowUA := strings.ToLower(ua)

	// Si el User-Agent contiene estas palabras, es bot SEGURO
	botKeywords := []string{
		"bot", "spider", "crawler", "facebookexternalhit", "facebot",
		"twitterbot", "linkedinbot", "whatsapp", "telegrambot",
		"google-layout-render", "headless",
	}

	for _, keyword := range botKeywords {
		if strings.Contains(lowUA, keyword) {
			return true
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true
	}

	for _, network := range compiledNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
