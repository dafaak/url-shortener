package utils

import (
	"net"
	"strings"
)

func IsBot(ua string, ipStr string) bool {
	lowUA := strings.ToLower(ua)
	
	
	botKeywords := []string{
		"bot", "spider", "crawler", "facebookexternalhit", 
		"linkedinbot", "twitterbot", "whatsapp", "telegrambot",
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

	// Lista de rangos conocidos de Datacenters/Redes Sociales
	networks := []string{
		"173.252.64.0/18",  // Facebook / Instagram
		"31.13.64.0/18",    // Facebook
		"108.174.0.0/20",   // LinkedIn
		"199.16.156.0/22",  // Twitter
		"34.192.0.0/10",    // AWS (Cubre muchos crawlers genéricos)
	}

	for _, cidr := range networks {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(ip) {
			return true
		}
	}

	return false
}