package app

import (
    "encoding/json"
    "fmt"
    "os"
    "runtime"
    "strings"

    "github.com/astaxie/beego"
    "github.com/astaxie/beego/cache"
    _ "github.com/astaxie/beego/cache/memcache"
    _ "github.com/astaxie/beego/cache/redis"
    "github.com/astaxie/beego/toolbox"

    _ "github.com/intelligide/off-api-proxy/internal/config"
)

var (
	Verbose bool = false
	Config AppConfig
	Cache cache.Cache
)

func Init() {
	if Config.CacheEnabled() {
		adapter := strings.Trim(Config.CacheAdapter(), " ")

		if len(adapter) > 0 {
			cacheCfg, err := Config.BConfig.DIY("cache." + adapter)
			var cacheCfgStr string
			if err == nil {
				b, err := json.Marshal(cacheCfg)
				if err != nil {
					//
				}
				cacheCfgStr = string(b)

			} else {
				cacheCfgStr = ""
			}
			c, err := cache.NewCache(adapter, cacheCfgStr)
			if err != nil {
				beego.Emergency("Cache:", err)
				os.Exit(1)
			}
			Cache = c
		}
	}

	memoryMonitorTask := toolbox.NewTask("memory_monitor", "0/5 * * * * *", memoryMonitor) // every 5s
	toolbox.AddTask("memory_monitor", memoryMonitorTask)
}

func memoryMonitor() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beego.Debug("Memory:", fmt.Sprintf("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC))

	if Config.CacheAdapter() == "memory" && uint64(Config.BConfig.DefaultInt64("cache.memory.max_memory", 100)) < bToMb(m.Alloc) {
		_ = Cache.ClearAll()
		runtime.GC()
	}

	return nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
