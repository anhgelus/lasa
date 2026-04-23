package config

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/redis/go-redis/v9"
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

func (c *Cache) Connect() (*redis.Client, error) {
	opts := redis.Options{
		Addr: c.Host + ":" + strconv.Itoa(int(c.Port)),
		DB:   int(c.DB),
	}
	if c.Auth != nil {
		if c.Auth.ClientName != "" {
			opts.ClientName = c.Auth.ClientName
		}
		if c.Auth.Username != "" {
			opts.Username = c.Auth.Username
		}
		opts.Password = c.Auth.Password
	}
	client := redis.NewClient(&opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client, client.Ping(ctx).Err()
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, toml.Unmarshal(b, &cfg)
}
