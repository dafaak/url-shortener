package utils

import (
	"net"
	"strings"
)

// Variable global para almacenar los rangos ya procesados en memoria
var compiledNetworks []*net.IPNet

// init se ejecuta automáticamente una sola vez al levantar el servidor
func init() {
	rawNetworks := []string{
		"173.252.64.0/18", // Meta
		"31.13.64.0/18",   // Meta
		"57.141.0.0/16",   // Meta (Aquí están las IPs 57.141.6.x que se filtraron)
		"108.174.0.0/20",  // LinkedIn
		"199.16.156.0/22", // X
		"34.192.0.0/10",   // AWS
		"54.144.0.0/12",   // AWS
	}

	for _, cidr := range rawNetworks {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			compiledNetworks = append(compiledNetworks, network)
		}
	}
}

// IsBot evalúa si una petición viene de un humano o un script
func IsBot(ua string, ipStr string) bool {
	// 1. Verificación por User-Agent (Keywords)
	lowUA := strings.ToLower(ua)
	botKeywords := []string{
		"bot", "spider", "crawler", "facebookexternalhit",
		"twitterbot", "linkedinbot", "whatsapp", "telegrambot",
	}

	for _, keyword := range botKeywords {
		if strings.Contains(lowUA, keyword) {
			return true
		}
	}

	// 2. Verificación por IP (Rangos de Datacenters)
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true // IP inválida = Petición sospechosa
	}

	// Iteramos sobre los rangos pre-compilados (muy rápido)
	for _, network := range compiledNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return false
}
