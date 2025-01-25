package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func initConfig() {
	// Set the config name and path
	viper.SetConfigName("config")   // name of config file (without extension)
	viper.SetConfigType("yaml")     // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("configs/") // path to look for the config file in
	viper.AddConfigPath(".")        // optionally look for config in the working directory

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found")
		} else {
			fmt.Printf("Error reading config: %v\n", err)
		}
		os.Exit(1)
	}

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.timeout", "30s")
}

func main() {

	// Initialize configuration
	initConfig()

	// Initialize Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", healthCheck)
	e.POST("/convert", convertHandler)

	// Start server
	e.Logger.Fatal(e.Start(":" + viper.GetString("server.port")))
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func convertHandler(c echo.Context) error {
	query := c.FormValue("query")
	// TODO: Add actual conversion logic
	return c.JSON(http.StatusOK, map[string]string{
		"query":  "rate(http_requests_total{status=~\"5..\"}[5m])",
		"input":  query,
		"status": "success",
	})
}
