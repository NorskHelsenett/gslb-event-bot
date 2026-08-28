package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var secretsKeyMap = map[string]string{
	"VK_USER":              "valkey.user",
	"VK_PASS":              "valkey.pass",
	"MQ_PASS":              "mq.pass",
	"MQ_USER":              "mq.user",
	"MQ_ENDPOINT":          "mq.endpoint",
	"SLACK_APP_TOKEN":      "slack.app_token",
	"SLACK_BOT_TOKEN":      "slack.bot_token",
	"SLACK_SIGNING_SECRET": "slack.signing_secret",
}

func new() (*Config, error) {
	v := viper.New()
	setDefaults(v)

	cfgPath := os.Getenv("CONFIG")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(cfgPath)
	v.AddConfigPath(".")
	v.AddConfigPath("/app")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config file: %w", err)
		}
		return nil, fmt.Errorf("un-expected error while loading configuration: %w", err)
	}

	loaded, total, err := loadSecrets(v, "./secrets")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	cfg.secretsLoaded = loaded
	cfg.secretsTotal = total

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.env", "prod")
	v.SetDefault("server.log_level", "INFO")

	v.SetDefault("valkey.addr", "localhost:6379")

	v.SetDefault("mq.endpoint", "localhost")
	v.SetDefault("mq.port", "5672")
}

func loadSecrets(v *viper.Viper, dir string) (loaded, total int, err error) {
	total = len(secretsKeyMap)

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return loaded, total, nil
	}

	if err != nil {
		return loaded, total, fmt.Errorf("failed to load secrets directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		key, ok := secretsKeyMap[entry.Name()]
		if !ok {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return loaded, total, fmt.Errorf("reading secret %s: %w", entry.Name(), err)
		}
		value := strings.TrimSpace(string(raw))
		if value != "" {
			v.Set(key, value)
			loaded++
		}
	}

	return loaded, total, nil

}
