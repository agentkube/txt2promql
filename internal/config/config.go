// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	AI         AIConfig         `mapstructure:"ai"`
	KG         KGConfig         `mapstructure:"knowledge_graph"`
	Semantic   SemanticConfig   `mapstructure:"semantic_memory"`
}

// ServerConfig
type ServerConfig struct {
	Port        int           `mapstructure:"port"`
	MaxBodySize string        `mapstructure:"max_body_size"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// PrometheusConfig
type PrometheusConfig struct {
	Address string        `mapstructure:"address"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// AI provider settings
type AIConfig struct {
	Model         string   `mapstructure:"model"`
	Temperature   float32  `mapstructure:"temperature"`
	TopP          float32  `mapstructure:"top_p"`
	BaseURL       string   `mapstructure:"base_url"`
	Proxy         string   `mapstructure:"proxy"`
	CustomHeaders []Header `mapstructure:"custom_headers"`
}

// HTTP header
type Header struct {
	Key   string `mapstructure:"key"`
	Value string `mapstructure:"value"`
}

// knowledge graph settings
type KGConfig struct {
	SchemaPath    string `mapstructure:"schema_path"`
	AutoDiscover  bool   `mapstructure:"auto_discover"`
	GraphEnabled  bool   `mapstructure:"enabled"`
	GraphURI      string `mapstructure:"graph_uri"`
	GraphUser     string `mapstructure:"graph_user"`
	GraphPassword string `mapstructure:"graph_password"`
	GraphDatabase string `mapstructure:"graph_database"`
}

// SemanticConfig holds semantic memory settings
type SemanticConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	FAISSIndex      string `mapstructure:"faiss_index"`
	EmbeddingsModel string `mapstructure:"embeddings_model"`
}

var (
	// Global configuration instance
	globalConfig *Config
)

func LoadConfig() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Load environment variables
	loadEnvVariables()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	globalConfig = config
	return config, nil
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	return globalConfig
}

// setDefaults sets default values for configuration
func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.max_body_size", "2MB")
	viper.SetDefault("server.timeout", "30s")

	// Prometheus defaults
	viper.SetDefault("prometheus.address", "http://localhost:9090")
	viper.SetDefault("prometheus.timeout", "30s")

	// AI defaults
	viper.SetDefault("ai.model", "gpt-4o-mini")
	viper.SetDefault("ai.temperature", 0.7)
	viper.SetDefault("ai.top_p", 1.0)

	// Knowledge Graph defaults
	viper.SetDefault("knowledge_graph.enabled", false)
	viper.SetDefault("knowledge_graph.auto_discover", true)
	viper.SetDefault("knowledge_graph.schema_path", "./schemas/prometheus.yaml")
	viper.SetDefault("knowledge_graph.graph_database", "neo4j")

	// Semantic Memory defaults
	viper.SetDefault("semantic_memory.enabled", true)
	viper.SetDefault("semantic_memory.faiss_index", "./data/faiss.index")
	viper.SetDefault("semantic_memory.embeddings_model", "sentence-transformers/all-MiniLM-L6-v2")
}

// loadEnvVariables loads environment variables into viper
func loadEnvVariables() {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		viper.Set("server.port", port)
	}

	if promURL := os.Getenv("PROMETHEUS_URL"); promURL != "" {
		viper.Set("prometheus.address", promURL)
	}

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		viper.Set("ai.api_key", apiKey)
	}

	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		viper.Set("ai.model", model)
	}

	if uri := os.Getenv("NEO4J_URI"); uri != "" {
		viper.Set("knowledge_graph.graph_uri", uri)
	}

	if user := os.Getenv("NEO4J_USER"); user != "" {
		viper.Set("knowledge_graph.graph_user", user)
	}

	if pass := os.Getenv("NEO4J_PASSWORD"); pass != "" {
		viper.Set("knowledge_graph.graph_password", pass)
	}
}

// validateConfig performs validation on the configuration
func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if cfg.Prometheus.Address == "" {
		return fmt.Errorf("prometheus address is required")
	}

	if cfg.AI.Temperature < 0 || cfg.AI.Temperature > 1 {
		return fmt.Errorf("AI temperature must be between 0 and 1")
	}
	if cfg.AI.TopP < 0 || cfg.AI.TopP > 1 {
		return fmt.Errorf("AI top_p must be between 0 and 1")
	}

	return nil
}
