package app

import (
	"github.com/astaxie/beego/config"
	"time"
)

type AppConfig struct {
	BConfig config.Configer
}

func (this *AppConfig) DataProvider() string {
	return this.BConfig.DefaultString("provider", "https://world.openfoodfacts.org/")
}

func (this *AppConfig) LogLevel() int {
	return this.BConfig.DefaultInt("logs.level", 6)
}

func (this *AppConfig) CacheEnabled() bool {
	return this.BConfig.DefaultBool("cache.enabled", true)
}

func (this *AppConfig) CacheAdapter() string {
	return this.BConfig.DefaultString("cache.adapter", "memory")
}

func (this *AppConfig) CacheTTL() time.Duration {
	d, _ := time.ParseDuration(this.BConfig.DefaultString("cache.ttl", "1h"))
	return d
}
