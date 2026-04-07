package config

import (
	"context"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
)

const DefaultPath = "/etc/lasad.toml"

type Config struct {
	Domain        string  `toml:"domain"`
	Port          uint    `toml:"port"`
	Cache         *Cache  `toml:"cache"`
	LegalNotice   *string `toml:"legal_notice_url"`
	LogNotFound   bool    `toml:"log_not_found"`
	LogBadRequest bool    `toml:"log_bad_request"`
}

type Cache struct {
	Host     string     `toml:"host"`
	Port     uint       `toml:"port"`
	DB       uint       `toml:"db"`
	Duration uint       `toml:"duration"`
	Auth     *CacheAuth `toml:"auth"`
}

type CacheAuth struct {
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	ClientName string `toml:"client_name"`
}

func (c *Cache) Connect() (*glide.Client, error) {
	cfg := config.NewClientConfiguration().
		WithAddress(&config.NodeAddress{Host: c.Host, Port: int(c.Port)})
	if c.Auth != nil {
		if c.Auth.Username != "" {
			cfg = cfg.WithCredentials(config.NewServerCredentials(c.Auth.Username, c.Auth.Password))
		} else {
			cfg = cfg.WithCredentials(config.NewServerCredentialsWithDefaultUsername(c.Auth.Password))
		}
	}
	client, err := glide.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.Ping(ctx)
	return client, err
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, toml.Unmarshal(b, &cfg)
}
