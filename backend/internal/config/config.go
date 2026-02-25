package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL             string `mapstructure:"url"`
	MaxOpenConns    int32  `mapstructure:"max_open_conns"`
	MaxIdleConns    int32  `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret             string        `mapstructure:"secret"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from config.yaml and environment variables.
// Environment variables override file values; they follow the format NEXUS_<KEY>.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AddConfigPath(".")

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "1h")
	v.SetDefault("jwt.access_token_expiry", "15m")
	v.SetDefault("jwt.refresh_token_expiry", "168h")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// Env overrides: NEXUS_DATABASE_URL, NEXUS_JWT_SECRET, etc.
	v.SetEnvPrefix("NEXUS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// Config file is optional; env vars are sufficient.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Required fields validation
	if cfg.Database.URL == "" {
		return nil, fmt.Errorf("database.url is required (set NEXUS_DATABASE_URL)")
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("jwt.secret is required (set NEXUS_JWT_SECRET)")
	}

	return &cfg, nil
}
