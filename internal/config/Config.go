package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl string	          `json:"db_url"`
	CurrentUserName string  `json:"current_user_name"`
}

func (c *Config)SetUser(userName string) error {
	c.CurrentUserName = userName
	return write(*c)
}

// Save persists the current configuration to disk
func (c *Config) Save() error {
	return write(*c)
} 

func Read() (Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	
	// Check if the config file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return Config{}, err
	}
	
	// Read the config file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}
	
	// Decode the JSON data into a Config struct
	var config Config
	err = json.Unmarshal(fileData, &config)
	if err != nil {
		return Config{}, err
	}
	
	return config, nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configFileName), nil
}

func write(cfg Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("error getting config file path: %w", err)
	}
	
	// Marshal the Config struct to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}
	
	// Write the JSON data to the config file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	
	return nil
}

