package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig 是全局应用配置。
var AppConfig Config

// Load reads config.yaml and .env, expands environment variables, and stores the
// result in AppConfig.
func Load() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := loadDotEnv(".env", v); err != nil {
		return err
	}

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	for _, key := range v.AllKeys() {
		if value, ok := v.Get(key).(string); ok {
			v.Set(key, expandEnv(value, v))
		}
	}

	return v.Unmarshal(&AppConfig)
}

func loadDotEnv(path string, v *viper.Viper) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		v.SetDefault(key, value)
	}

	return scanner.Err()
}

func expandEnv(value string, v *viper.Viper) string {
	return os.Expand(value, func(key string) string {
		if v.IsSet(key) {
			return v.GetString(key)
		}
		return os.Getenv(key)
	})
}
