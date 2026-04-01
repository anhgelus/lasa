package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/valkey-io/valkey-go"
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

func (c *Cache) Connect() (valkey.Client, error) {
	addr := c.Host
	if c.Port > 0 {
		addr += fmt.Sprintf(":%d", c.Port)
	}
	return valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
		SelectDB:    int(c.DB),
		Username:    c.Auth.Username,
		Password:    c.Auth.Password,
		ClientName:  c.Auth.ClientName,
	})
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, toml.Unmarshal(b, cfg)
}
