package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

var localConfig = ".jprq-config"
var remoteConfig = "https://jprq.io/config.json"

type Config struct {
	Remote struct {
		Domain string `json:"domain"`
		Events string `json:"events"`
	}
	Local struct {
		AuthToken string `json:"auth_token"`
	}
}

func (c *Config) Load() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %s", err)
	}
	filePath := filepath.Join(homeDir, localConfig)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	if err := json.Unmarshal(data, &c.Local); err != nil {
		return fmt.Errorf("error unmarshaling config file contents: %s", err)
	}
	response, err := http.Get(remoteConfig)
	if err != nil || response.StatusCode != http.StatusOK {
		return fmt.Errorf("error fetching %s: %s", remoteConfig, err)
	}
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&c.Remote); err != nil {
		return fmt.Errorf("error decoding config file: %s", err)
	}
	return nil
}

func (c *Config) Write() error {
	content, err := json.Marshal(c.Local)
	if err != nil {
		return fmt.Errorf("error marshaling config: %s", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %s", err)
	}
	filePath := filepath.Join(homeDir, localConfig)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating config file: %s", err)
	}
	if _, err = file.Write(content); err != nil {
		return fmt.Errorf("error writitng to config file: %s", err)
	}
	return nil
}
