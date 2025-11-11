package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"wsun.dev/porkbun-dns-updater/porkbun"
)

func main() {
	// Setup logger
	log.SetOutput(os.Stdout)

	// Parse env variables
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	domain := os.Getenv("DOMAIN")
	subdomainsStr := os.Getenv("SUBDOMAINS")
	ttl := os.Getenv("TTL")
	if ttl == "" {
		ttl = "600"
	}

	if apiKey == "" || secretKey == "" || domain == "" {
		log.Fatal("API_KEY, SECRET_KEY, and DOMAIN are required")
	}

	client := porkbun.New(apiKey, secretKey, domain, ttl)

	subdomains := strings.Split(subdomainsStr, ",")
	if subdomainsStr == "" {
		subdomains = []string{""} // Default to root domain
	}

	ip, err := client.GetIP()
	if err != nil {
		log.Fatalf("Failed to get public IP: %v", err)
	}

	for _, sub := range subdomains {
		subdomain := strings.TrimSpace(sub)
		fullDomain := domain
		if subdomain != "" {
			fullDomain = fmt.Sprintf("%s.%s", subdomain, domain)
		}
		log.Printf("Updating porkbun for %s", fullDomain)
		if err := client.UpsertDNSRecord(ip, subdomain); err != nil {
			log.Fatalf("Failed to upsert DNS record for %s: %v", fullDomain, err)
		}
	}
	log.Printf("DNS update/check complete for all subdomains. ip=%s", ip)
}
