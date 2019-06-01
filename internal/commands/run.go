package commands

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/toolbox"
	"github.com/intelligide/off-api-proxy/internal/config"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Long:  ``,
	Run: exec,
}

var (
	verbose bool = false
	_cache *cache.Cache

)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&config.GlobalConfig.Http.Address, "address", config.GlobalConfig.Http.Address, "Address")
	runCmd.Flags().IntVarP(&config.GlobalConfig.Http.Port, "port", "p", config.GlobalConfig.Http.Port, "Port")
	runCmd.Flags().BoolVarP(&verbose, "verbose",  "v", verbose, "Verbose")
}

func exec(cmd *cobra.Command, args []string) {
	if config.GlobalConfig.Cache.Enabled {
		_cache = cache.New(time.Duration(config.GlobalConfig.Cache.Expiration) * time.Minute, time.Duration(config.GlobalConfig.Cache.CleanupInterval) * time.Minute);
	}

	if verbose {
		beego.SetLevel(beego.LevelDebug)
	} else {
		beego.SetLevel(config.GlobalConfig.Logs.Level)
	}

	config.GlobalConfig.ConfigureBeego(beego.BConfig)

	memoryMonitorTask := toolbox.NewTask("memory_monitor", "0/5 * * * * *", memoryMonitor) // every 5s
	toolbox.AddTask("memory_monitor", memoryMonitorTask)
	toolbox.StartTask()
	defer toolbox.StopTask()

	beego.Get("/api/v0/product/:product_id:int.json", ProxyFunc)
	beego.Get("/api/v0/product/batch.json", Batch)
	beego.Run()
}

func ProxyFunc(ctx *context.Context) {

	product_id := ctx.Input.Param(":product_id")

	if config.GlobalConfig.Cache.Enabled {
		product, inCache := _cache.Get(product_id)
		if(inCache) {
			beego.Debug("Fetch product " + product_id + " from cache")
			ctx.Output.JSON(product, false, true)
			return
		}
	}

	if ctx.Input.Context.Request.Form == nil {
		ctx.Input.Context.Request.ParseForm()
	}


	provider := config.GlobalConfig.Provider
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


	if int(dat["status"].(float64)) == 1 && config.GlobalConfig.Cache.Enabled {
		_cache.Add(product_id, dat, cache.DefaultExpiration)
	}

	ctx.Output.JSON(dat, false, true)


}

func Batch(ctx *context.Context) {

}

func memoryMonitor() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	beego.Debug(fmt.Sprintf("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC))

	if(uint64(config.GlobalConfig.Cache.MaxAllocMemory) < bToMb(m.Alloc)) {
		_cache.Flush()
		runtime.GC()
	}

	return nil
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}