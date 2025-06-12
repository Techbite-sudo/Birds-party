package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration from environment
type Config struct {
	RNGServiceURL      string
	SettingsServiceURL string
	ServerPort         string
	LogFile            string
}

// Load loads configuration from environment variables
func Load() Config {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	return Config{
		RNGServiceURL:      getEnv("RNG_API_URL", "http://159.89.235.166:17003/api/proxy/rng/1"),
		SettingsServiceURL: getEnv("SETTINGS_API_URL", "https://t3.ibibe.africa/get-game-settings"),
		ServerPort:         getEnv("PORT", "11400"),
		LogFile:            getEnv("LOG_FILE", "app.log"),
	}
}

// Function to get an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// LoadAll loads both production and test configurations from environment variables
func LoadAll() (prod Config, test Config) {
	// Try to load .env file, but don't fail if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	prod = Config{
		RNGServiceURL:      getEnv("PROD_RNG_API_URL", "http://159.89.235.166:17003/api/proxy/rng/1"),
		SettingsServiceURL: getEnv("PROD_SETTINGS_API_URL", "https://t3.ibibe.africa/get-game-settings"),
		ServerPort:         getEnv("PORT", "11400"),
		LogFile:            getEnv("LOG_FILE", "app.log"),
	}
	test = Config{
		RNGServiceURL:      getEnv("TEST_RNG_API_URL", "http://test-rng-url"),
		SettingsServiceURL: getEnv("TEST_SETTINGS_API_URL", "https://test-settings-url"),
		ServerPort:         getEnv("PORT", "11400"),
		LogFile:            getEnv("LOG_FILE", "app.log"),
	}
	return
}
