package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type RepositoryConfig struct {
	//Path 仓库路径
	Path string `json:"path"`
	//Username 用户名
	Username string `json:"username"`
	//User git用户名，默认为 git
	User string `json:"user"`
	//PrivateKeyFile 密钥路径，默认为 ~/.ssh/id_rsa
	PrivateKeyFile string `json:"private_key_file"`
	//Password 密码
	Password string `json:"password"`
	//Interval 同步间隔，默认为60秒
	Interval int `json:"interval"` // 扫描间隔，单位:秒
	//IncludePatterns 只同步指定规则的文件，如果为空则同步所有文件
	IncludePatterns []string `json:"include_patterns"`
	//ExcludePatterns 需要排除的指定规则文件，凡是匹配的都不会同步
	ExcludePatterns []string `json:"exclude_patterns"`
}

type Repository struct {
	authMethod     transport.AuthMethod
	logger         *log.Logger
	skipEmptyFiles bool
	gitRepository  *git.Repository
	Path           string
	Interval       int
	includeRegexps []*regexp.Regexp
	excludeRegexps []*regexp.Regexp
}

func NewRepository(cfg RepositoryConfig) (*Repository, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("path is required")
	}
	stat, err := os.Stat(cfg.Path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%w", err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not directory", cfg.Path)
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 60
	}
	gitRepo, err := git.PlainOpen(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open git repository failed: %w", err)
	}
	repo := Repository{
		logger:         log.New(os.Stdout, "[GitSync]", log.LstdFlags),
		gitRepository:  gitRepo,
		Path:           cfg.Path,
		Interval:       cfg.Interval,
		includeRegexps: make([]*regexp.Regexp, 0),
		excludeRegexps: make([]*regexp.Regexp, 0),
	}
	if len(cfg.Username) > 0 {
		repo.authMethod = &http.BasicAuth{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}
	if len(cfg.PrivateKeyFile) > 0 {
		repo.authMethod, err = ssh.NewPublicKeysFromFile(cfg.User, cfg.PrivateKeyFile, cfg.Password)
		if err != nil {
			return nil, fmt.Errorf("generate publickeys failed: %w", err)
		}
	}
	for _, pattern := range cfg.IncludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("parse regexp pattern failed: %w\n", err)
		}
		repo.includeRegexps = append(repo.includeRegexps, re)
	}
	for _, pattern := range cfg.ExcludePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("parse regexp pattern failed: %w\n", err)
		}
		repo.excludeRegexps = append(repo.excludeRegexps, re)
	}
	return &repo, nil
}

func (repo *Repository) Start() {
	for {
		worktree, err := repo.gitRepository.Worktree()
		if err != nil {
			repo.logger.Printf("get worktree failed: %v\n", err)
			continue
		}
		status, err := worktree.Status()
		if err != nil {
			repo.logger.Printf("get status failed: %v\n", err)
		}
		c := 0
		for path := range status {
			if repo.isBlocked(path) {
				continue
			}
			if repo.isMatched(path) {
				if _, err := worktree.Add(path); err != nil {
					continue
				}
				c++
			}
		}
		if c == 0 {
			continue
		}
		_, err = worktree.Commit("auto commit", &git.CommitOptions{})
		if err != nil {
			repo.logger.Printf("commit failed: %v\n", err)
		}
		repo.logger.Printf("commit succeeded\n")
		remotes, err := repo.gitRepository.Remotes()
		if err != nil {
			repo.logger.Printf("get remotes failed: %v\n", err)
			continue
		}
		if len(remotes) > 0 {
			if err := worktree.Pull(&git.PullOptions{
				Auth: repo.authMethod,
			}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
				repo.logger.Printf("pull failed: %v\n", err)
			}
			if err := repo.gitRepository.Push(&git.PushOptions{
				Auth: repo.authMethod,
			}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
				repo.logger.Printf("push failed: %v\n", err)
			}
		}
		time.Sleep(time.Duration(repo.Interval) * time.Second)
	}
}

func (repo Repository) isMatched(path string) bool {
	if len(repo.includeRegexps) == 0 {
		return true
	}
	for _, re := range repo.includeRegexps {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

func (repo Repository) isBlocked(path string) bool {
	for _, re := range repo.excludeRegexps {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}
