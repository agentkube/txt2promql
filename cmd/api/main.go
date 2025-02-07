package main

import (
	"context"
	"fmt"
	"os"

	"github.com/agentkube/txt2promql/internal/config"
	"github.com/agentkube/txt2promql/internal/core/knowledgegraph"
	"github.com/agentkube/txt2promql/internal/prometheus"
	"github.com/agentkube/txt2promql/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found")
		} else {
			fmt.Printf("Error reading config: %v\n", err)
		}
		os.Exit(1)
	}

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.timeout", "30s")
}

func main() {
	initConfig()

	// Initialize Prometheus client
	promClient := prometheus.NewClient()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var kgClient *knowledgegraph.Client
	if cfg.KG.GraphEnabled {
		kgClient, err = knowledgegraph.NewClient(&knowledgegraph.Config{
			Enabled:  true,
			URI:      cfg.KG.GraphURI,
			User:     cfg.KG.GraphUser,
			Password: cfg.KG.GraphPassword,
			Database: cfg.KG.GraphDatabase,
		})
		if err != nil {
			fmt.Printf("Warning: Failed to initialize knowledge graph: %v\n", err)
			// Continue without knowledge graph
		} else {
			// Test connection
			if err := kgClient.Connect(context.Background()); err != nil {
				fmt.Printf("Warning: Failed to connect to knowledge graph: %v\n", err)
				kgClient = nil // Disable knowledge graph functionality
			} else {
				fmt.Println("Successfully connected to knowledge graph")
			}
		}
	}

	// Initialize Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register handlers
	if err := server.RegisterHandlers(e, promClient, kgClient); err != nil {
		fmt.Printf("Error registering handlers: %v\n", err)
		os.Exit(1)
	}
	// Start server
	e.Logger.Fatal(e.Start(":" + viper.GetString("server.port")))
}
