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
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("error getting user config directory: %s", err)
	}
	filePath := filepath.Join(configDir, "jprq", localConfig)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error: no auth token, obtain at https://jprq.io/auth")
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
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("error getting user config directory: %s", err)
	}
	dirPath := filepath.Join(configDir, "jprq")
    if err := os.MkdirAll(dirPath, 0700); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("error creating config directory: %s", err)
    }
    filePath := filepath.Join(dirPath, localConfig)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating config file: %s", err)
	}
	if _, err = file.Write(content); err != nil {
		return fmt.Errorf("error writitng to config file: %s", err)
	}
	return nil
}
