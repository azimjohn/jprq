package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"github.com/azimjohn/jprq/cli/i18n"
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
		Language string `json:"language"`
	}
}

func (c *Config) Load() error {

	configDir, err := os.UserConfigDir()

	if err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_getting_user_config"), err)
	}

	filePath := filepath.Join(configDir, "jprq", localConfig)
	data, err := os.ReadFile(filePath)

	if err != nil {
		return fmt.Errorf("%s https://jprq.io/auth", i18n.T("no_auth_token"))
	}

	if err := json.Unmarshal(data, &c.Local); err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_unmarshaling"), err)
	}

	if c.Local.Language == "" {
    c.Local.Language = "en"
}

	response, err := http.Get(remoteConfig)

	if err != nil || response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: %s", i18n.T("error_fetching"), remoteConfig, err)
	}
	
	defer response.Body.Close()

	if err := json.NewDecoder(response.Body).Decode(&c.Remote); err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_decoding_config"), err)
	}

	return nil
}

func (c *Config) Write() error {
	content, err := json.Marshal(c.Local)
	if err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_marshaling_config"), err)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_getting_user_config"), err)
	}
	dirPath := filepath.Join(configDir, "jprq")
	if err := os.MkdirAll(dirPath, 0700); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("%s %s", i18n.T("error_creating_config_directory"), err)
	}
	filePath := filepath.Join(dirPath, localConfig)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_creating_config_file"), err)
	}
	if _, err = file.Write(content); err != nil {
		return fmt.Errorf("%s %s", i18n.T("error_writing_config_file"), err)
	}
	return nil
}
