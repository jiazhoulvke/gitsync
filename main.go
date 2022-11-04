package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/pflag"
)

var (
	cfgFile string
	cfg     Config
)

func init() {
	pflag.StringVarP(&cfgFile, "config", "c", "gitsync.json", "config file path")
}

func main() {
	pflag.Parse()
	cfg, err := LoadConfig(cfgFile)
	if err != nil {
		fmt.Printf("load config failed: %v", err)
		os.Exit(1)
	}
	if len(cfg.Repos) == 0 {
		fmt.Println("no repos found")
		os.Exit(1)
	}
	var wg sync.WaitGroup
	for _, repoCfg := range cfg.Repos {
		repo, err := NewRepository(repoCfg)
		if err != nil {
			fmt.Printf("initialize repo %s failed: %v\n", repoCfg.Path, err)
			os.Exit(1)
		}
		wg.Add(1)
		fmt.Printf("gitsync start at %s\n", repoCfg.Path)
		go repo.Start()
	}
	wg.Wait()
}
