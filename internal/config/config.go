package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Jobs     JobsConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port            string
	Env             string
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	MaxOpen  int
	MaxIdle  int
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type JobsConfig struct {
	Workers    int
	QueueSize  int
}

type UploadConfig struct {
	Dir       string
	MaxSize   int64
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
