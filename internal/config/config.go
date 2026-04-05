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
	CORS     CORSConfig
}

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	Env             string        `mapstructure:"env"`
	ShutdownTimeout time.Duration `mapstructure:"shutdowntimeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"dbport"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	MaxOpen  int    `mapstructure:"maxopen"`
	MaxIdle  int    `mapstructure:"maxidle"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	Expiry        time.Duration `mapstructure:"expiry"`
	RefreshExpiry time.Duration `mapstructure:"refreshexpiry"`
}

type JobsConfig struct {
	Workers   int `mapstructure:"workers"`
	QueueSize int `mapstructure:"queuesize"`
}

type UploadConfig struct {
	Dir     string `mapstructure:"dir"`
	MaxSize int64  `mapstructure:"maxsize"`
}

type CORSConfig struct {
	AllowedOrigins   []string      `mapstructure:"allowedorigins"`
	AllowedMethods   []string      `mapstructure:"allowedmethods"`
	AllowedHeaders   []string      `mapstructure:"allowedheaders"`
	AllowCredentials bool          `mapstructure:"allowcredentials"`
	MaxAge           time.Duration `mapstructure:"maxage"`
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
