package config

import (
	"testing"
)

func TestConfig_Load(t *testing.T) {
	t.Setenv("JPRQ_DOMAIN", "jprq.live")
	t.Setenv("JPRQ_TLS_KEY", "key.pem")
	t.Setenv("JPRQ_TLS_CERT", "cert.pem")

	config := &Config{}
	err := config.Load()
	if err != nil {
		t.Logf("Error while loading the config: %v", err.Error())
		t.Fail()
	}
}

func TestConfig_LoadEmptyEnv(t *testing.T) {
	cases := []struct {
		description string
		domainEnv   string
		keyEnv      string
		certEnv     string
		expected    string
	}{
		{
			description: "JPRQ_DOMAIN env is not provided",
			domainEnv:   "",
			keyEnv:      "example.key",
			certEnv:     "example.cert",
			expected:    "JPRQ_DOMAIN env is not set",
		},
		{
			description: "JPRQ_TLS_KEY env is not provided",
			domainEnv:   "jprq.live",
			keyEnv:      "",
			certEnv:     "example.cert",
			expected:    "TLS key/cert file is missing",
		},
		{
			description: "JPRQ_TLS_CERT not is not provided",
			domainEnv:   "jprq.live",
			keyEnv:      "example.key",
			certEnv:     "",
			expected:    "TLS key/cert file is missing",
		},
	}
	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			t.Setenv("JPRQ_DOMAIN", tt.domainEnv)
			t.Setenv("JPRQ_TLS_KEY", tt.keyEnv)
			t.Setenv("JPRQ_TLS_CERT", tt.certEnv)

			config := &Config{}
			err := config.Load()
			if err == nil {
				t.Logf("expected %v, but got %v", tt.expected, err)
			}
			if err.Error() != tt.expected {
				t.Logf("expected %s, but got %s", tt.expected, err.Error())
			}
		})
	}
}
