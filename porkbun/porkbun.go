package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	apiBaseURL = "https://api-ipv4.porkbun.com/api/json/v3/"
)

type Client struct {
	APIKey    string
	SecretKey string
	Domain    string
	TTL       string
}

type APIRequest struct {
	APIKey       string `json:"apikey"`
	SecretAPIKey string `json:"secretapikey"`
	Content      string `json:"content,omitempty"`
	TTL          string `json:"ttl,omitempty"`
	Type         string `json:"type,omitempty"`
	Name         string `json:"name,omitempty"`
}

type PingResponse struct {
	Status string `json:"status"`
	YourIP string `json:"yourIp"`
}

type GetRecordsResponse struct {
	Status  string   `json:"status"`
	Records []Record `json:"records"`
}

type Record struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     string `json:"ttl"`
	Notes   string `json:"notes"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

func New(apiKey, secretKey, domain, ttl string) *Client {
	return &Client{
		APIKey:    apiKey,
		SecretKey: secretKey,
		Domain:    domain,
		TTL:       ttl,
	}
}

func (c *Client) post(path string, body interface{}) ([]byte, error) {
	url := apiBaseURL + path
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) GetIP() (string, error) {
	reqBody := APIRequest{
		APIKey:       c.APIKey,
		SecretAPIKey: c.SecretKey,
	}
	respBody, err := c.post("ping", reqBody)
	if err != nil {
		return "", err
	}

	var pingResp PingResponse
	if err := json.Unmarshal(respBody, &pingResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal ping response: %w", err)
	}

	if pingResp.Status != "SUCCESS" {
		return "", fmt.Errorf("ping request was not successful: %s", string(respBody))
	}

	return pingResp.YourIP, nil
}

func (c *Client) GetDNSRecords(subdomain string) ([]Record, error) {
	path := fmt.Sprintf("dns/retrieveByNameType/%s/A/", c.Domain)
	if subdomain != "" {
		path = fmt.Sprintf("dns/retrieveByNameType/%s/A/%s", c.Domain, subdomain)
	}

	reqBody := APIRequest{
		APIKey:       c.APIKey,
		SecretAPIKey: c.SecretKey,
	}
	respBody, err := c.post(path, reqBody)
	if err != nil {
		return nil, err
	}

	var recordsResp GetRecordsResponse
	if err := json.Unmarshal(respBody, &recordsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal records response: %w", err)
	}

	if recordsResp.Status != "SUCCESS" {
		return nil, fmt.Errorf("get records request was not successful: %s", string(respBody))
	}

	return recordsResp.Records, nil
}

func (c *Client) CreateDNSRecord(ip, subdomain string) error {
	log.Printf("Creating DNS record for subdomain '%s' with IP %s", subdomain, ip)
	path := fmt.Sprintf("dns/create/%s", c.Domain)
	reqBody := APIRequest{
		APIKey:       c.APIKey,
		SecretAPIKey: c.SecretKey,
		Type:         "A",
		Name:         subdomain,
		Content:      ip,
		TTL:          c.TTL,
	}

	respBody, err := c.post(path, reqBody)
	if err != nil {
		return err
	}

	var statusResp StatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return fmt.Errorf("failed to unmarshal create response: %w", err)
	}

	if statusResp.Status != "SUCCESS" {
		return fmt.Errorf("create record request was not successful: %s", string(respBody))
	}
	log.Printf("Created dns record with ip %s", ip)
	return nil
}

func (c *Client) UpdateDNSRecord(ip, subdomain string) error {
	log.Printf("Updating DNS record for subdomain '%s' to IP %s", subdomain, ip)
	path := fmt.Sprintf("dns/editByNameType/%s/A/", c.Domain)
	if subdomain != "" {
		path = fmt.Sprintf("dns/editByNameType/%s/A/%s", c.Domain, subdomain)
	}

	reqBody := APIRequest{
		APIKey:       c.APIKey,
		SecretAPIKey: c.SecretKey,
		Content:      ip,
		TTL:          c.TTL,
	}

	respBody, err := c.post(path, reqBody)
	if err != nil {
		return err
	}

	var statusResp StatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return fmt.Errorf("failed to unmarshal update response: %w", err)
	}

	if statusResp.Status != "SUCCESS" {
		return fmt.Errorf("update record request was not successful: %s", string(respBody))
	}
	log.Printf("Updated dns record with ip %s", ip)
	return nil
}

func (c *Client) UpsertDNSRecord(ip, subdomain string) error {
	records, err := c.GetDNSRecords(subdomain)
	if err != nil {
		return fmt.Errorf("failed to get DNS records: %w", err)
	}

	if len(records) > 1 {
		return fmt.Errorf("could not update DNS record as multiple records were found for subdomain '%s'", subdomain)
	}

	if len(records) == 0 {
		return c.CreateDNSRecord(ip, subdomain)
	}

	record := records[0]
	if record.Content == ip {
		log.Printf("Skipping update for subdomain '%s' as record already matches current ip %s", subdomain, ip)
		return nil
	}

	return c.UpdateDNSRecord(ip, subdomain)
}
