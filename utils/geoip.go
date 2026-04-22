package utils

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

func GetCountryFromIP(ipAddr string, dbPath string) string {
	// 1. Abrir la base de datos
	db, err := geoip2.Open(dbPath)
	if err != nil {
		log.Printf("ERROR: No se encontró el archivo MMDB en: %s", dbPath)
		return "Unknown"
	}
	defer db.Close()

	// 2. Parsear la IP
	ip := net.ParseIP(ipAddr)

	// 3. Buscar el registro
	record, err := db.Country(ip)
	if err != nil {
		return "Unknown"
	}

	// Retorna el nombre en español si existe, si no en inglés
	if name, ok := record.Country.Names["es"]; ok {
		return name
	}
	return record.Country.IsoCode
}
