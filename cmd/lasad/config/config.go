package config

import (
	"os"

	"github.com/BurntSushi/toml"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
)

const DefaultPath = "/etc/lasad.toml"

type Config struct {
	Domain string `toml:"domain"`
	Port   uint   `toml:"port"`
	Cache  *Cache `toml:"cache"`
}

type Cache struct {
	Host     string    `toml:"host"`
	Port     uint      `toml:"port"`
	DB       uint      `toml:"db"`
	Duration uint      `toml:"duration"`
	Auth     CacheAuth `toml:"auth"`
}

type CacheAuth struct {
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	ClientName string `toml:"client_name"`
}

func (c *Cache) Connect() (*glide.Client, error) {
	cfg := config.NewClientConfiguration().
		WithAddress(&config.NodeAddress{Host: c.Host, Port: int(c.Port)})
	if c.Auth.Password != "" {
		if c.Auth.Username != "" {
			cfg = cfg.WithCredentials(config.NewServerCredentials(c.Auth.Username, c.Auth.Password))
		} else {
			cfg = cfg.WithCredentials(config.NewServerCredentialsWithDefaultUsername(c.Auth.Password))
		}
	}
	return glide.NewClient(cfg)
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, toml.Unmarshal(b, &cfg)
}
