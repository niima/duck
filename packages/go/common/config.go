package common

import "os"

// Config holds common configuration
type Config struct {
	AppName string
	Port    string
}

// NewConfig creates a new config instance
func NewConfig(appName string) *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		AppName: appName,
		Port:    port,
	}
}
