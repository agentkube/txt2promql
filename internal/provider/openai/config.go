package openai

import (
	"net/http"
)

// IAIConfig interface defines configuration methods for AI providers
type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
	GetProxyEndpoint() string
	GetOrganizationId() string
	GetTemperature() float32
	GetTopP() float32
	GetCustomHeaders() []http.Header
}

// Config implements IAIConfig interface
type Config struct {
	APIKey        string              `mapstructure:"api_key"`
	Model         string              `mapstructure:"model"`
	BaseURL       string              `mapstructure:"base_url"`
	ProxyEndpoint string              `mapstructure:"proxy_endpoint"`
	OrgID         string              `mapstructure:"org_id"`
	Temperature   float32             `mapstructure:"temperature"`
	TopP          float32             `mapstructure:"top_p"`
	CustomHeaders map[string][]string `mapstructure:"custom_headers"`
}

func (c *Config) GetPassword() string       { return c.APIKey }
func (c *Config) GetModel() string          { return c.Model }
func (c *Config) GetBaseURL() string        { return c.BaseURL }
func (c *Config) GetProxyEndpoint() string  { return c.ProxyEndpoint }
func (c *Config) GetOrganizationId() string { return c.OrgID }
func (c *Config) GetTemperature() float32   { return c.Temperature }
func (c *Config) GetTopP() float32          { return c.TopP }
func (c *Config) GetCustomHeaders() []http.Header {
	var headers []http.Header
	for key, values := range c.CustomHeaders {
		header := make(http.Header)
		for _, value := range values {
			header.Add(key, value)
		}
		headers = append(headers, header)
	}
	return headers
}
