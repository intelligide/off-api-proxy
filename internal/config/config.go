package config

import (
	"github.com/astaxie/beego"
	"github.com/pelletier/go-toml"
)

type Config struct {
	Cache CacheConfig `toml:"cache"`
	Logs LogConfig `toml:"logs"`
	Provider string `default:"https://world.openfoodfacts.org/" toml:"provider"`
	PreFilters []string `toml:"pre-filters"`
	Filters []string `toml:"filters"`
}

type CacheConfig struct {
	Enabled bool `default:"true" toml:"enabled"` // minutes
	Expiration uint `default:"60" toml:"expires_after"` // minutes
	CleanupInterval uint `default:"20" toml:"cleanup_interval"` // minutes
	MaxAllocMemory uint `default:"100" toml:"max_memory"` // MiB
}

type LogConfig struct {
	Level int `default:"5" toml:"level"`
}

var (
	GlobalConfig *Config
)

func init() {
	GlobalConfig = newConfig()
	t, err := toml.LoadFile("config.toml")
	if err != nil {
		beego.Info("Cannot read config.tml. Using default values.")
	} else {
		t.Unmarshal(GlobalConfig)
	}
}

func newConfig() *Config {
	return &Config {
		Cache: CacheConfig{
			Enabled: true,
			Expiration: 60,
			CleanupInterval: 20,
			MaxAllocMemory: 100,
		},
		Logs: LogConfig{
			Level: 5,
		},
		Provider: "https://world.openfoodfacts.org/",
	}
}
