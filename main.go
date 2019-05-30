package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/toolbox"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"net/url"
	"openfoodfacts-proxy/internal/config"
	"path"
	"runtime"
	"time"
)

var _cache *cache.Cache
var _config *config.Config

func ProxyFunc(ctx *context.Context) {

	product_id := ctx.Input.Param(":product_id")

	product, inCache := _cache.Get(product_id)
	if(inCache) {
		beego.Debug("Fetch product " + product_id + " from cache")
		ctx.Output.JSON(product, false, true)
	} else {
		if ctx.Input.Context.Request.Form == nil {
			ctx.Input.Context.Request.ParseForm()
		}


		provider := "https://world.openfoodfacts.org/"

		u, err := url.Parse(provider)
		if err != nil {
			panic(err)
		}

		u.Path = path.Join(u.Path, "/api/v0/product/" + ctx.Input.Param(":product_id") +".json")

		q := ctx.Input.Context.Request.Form
		delete(q, "filters")
		u.RawQuery = q.Encode()
		urlstring := u.String()

		beego.Debug("Fetch product " + product_id + " from " + provider + "(" + urlstring + ")")
		resp, err := http.Get(urlstring)
		if err != nil {
			beego.Error(err)
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if  err != nil {
			beego.Error(err)
			return
		}

		var dat map[string]interface{}

		if err := json.Unmarshal(body, &dat); err != nil {
			beego.Error(err)
			return
		}


		if int(dat["status"].(float64)) == 1 {
			_cache.Add(product_id, dat, cache.DefaultExpiration)
		}

		ctx.Output.JSON(dat, false, true)

	}
}

func memoryMonitor() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)

	/*
	if(uint64(_config.MaxAllocMemory) < bToMb(m.Alloc)) {
		_cache.Flush()
		 runtime.GC()
	}*/

	return nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func main() {
	_config := config.Load()

	_cache = cache.New(time.Duration(_config.Expiration) * time.Minute, time.Duration(_config.CleanupInterval) * time.Minute);

	memoryMonitorTask := toolbox.NewTask("memory_monitor", "0/5 * * * * *", memoryMonitor)
	toolbox.AddTask("memory_monitor", memoryMonitorTask)
	toolbox.StartTask()
	defer toolbox.StopTask()

	beego.Get("/api/v0/product/:product_id:int.json", ProxyFunc)
	beego.Run()
}

