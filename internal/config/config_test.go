package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoad_FileNotFound(t *testing.T) {
	v := viper.New()
	v.SetConfigFile("/nonexistent/path/.env")
	v.SetConfigType("env")

	err := v.ReadInConfig()
	assert.Error(t, err)
}

func TestLoad_WithValidEnvFile(t *testing.T) {
	content := `PORT=8080
ENV=test
`
	tmpFile, err := os.CreateTemp("", "config-*.env")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)
	tmpFile.Close()

	v := viper.New()
	v.SetConfigFile(tmpFile.Name())
	v.SetConfigType("env")

	err = v.ReadInConfig()
	assert.NoError(t, err)

	var cfg Config
	err = v.Unmarshal(&cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestConfig_StructDefaults(t *testing.T) {
	cfg := Config{}

	assert.Equal(t, "", cfg.Server.Port)
	assert.Equal(t, "", cfg.Server.Env)
	assert.Equal(t, "", cfg.Database.Host)
	assert.Equal(t, "", cfg.JWT.Secret)
	assert.Equal(t, "", cfg.Upload.Dir)
	assert.Equal(t, int64(0), cfg.Upload.MaxSize)
}
