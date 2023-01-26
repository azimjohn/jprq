package config

import (
	"fmt"
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

func TestConfig_loadEmptvEnv2(t *testing.T) {
	envs := []struct {
		key     string
		value   string
		ErrText string
	}{
		{
			"JPRQ_DOMAIN",
			"jprq.live",
			"JPRQ_DOMAIN env is not set",
		},
		{
			"JPRQ_TLS_KEY",
			"example.key",
			"TLS key/cert file is missing",
		},
		{
			"JPRQ_TLS_CERT",
			"example.cert",
			"TLS key/cert file is missing",
		},
	}

	for i, missing := range envs {
		t.Run(fmt.Sprintf("Missing %s", missing.key), func(t *testing.T) {
			for j, env := range envs {
				if i == j {
					continue
				}
				t.Setenv(env.key, env.value)
			}
			config := &Config{}
			err := config.Load()

			if err == nil {
				t.Logf("expected %v, but got %v", missing.ErrText, err)
			}
			if err.Error() != missing.ErrText {
				t.Logf("expected %s, but got %s", missing.ErrText, err.Error())
			}
		})
	}
}
