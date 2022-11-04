package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/muhammadmuzzammil1998/jsonc"
)

type Config struct {
	Repos []RepositoryConfig `json:"repos"`
}

func LoadConfig(configFile string) (*Config, error) {
	if configFile == "" {
		return nil, errors.New("config file not found")
	}
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %w", err)
	}
	var cfg Config
	if err := jsonc.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file failed: %w", err)
	}
	for i, repo := range cfg.Repos {
		if repo.Path == "" {
			return nil, errors.New("git repository path could not be empty")
		}
		cfg.Repos[i].Path = AbsPath(repo.Path)
		if repo.Interval <= 0 {
			cfg.Repos[i].Interval = 60
		}
		if repo.PrivateKeyFile == "" {
			cfg.Repos[i].PrivateKeyFile = "~/.ssh/id_rsa"
		}
		cfg.Repos[i].PrivateKeyFile = AbsPath(repo.PrivateKeyFile)
		if repo.User == "" {
			cfg.Repos[i].User = "git"
		}
	}
	return &cfg, nil
}
