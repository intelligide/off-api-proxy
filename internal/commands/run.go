package commands

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
    "github.com/astaxie/beego/config"
    "github.com/spf13/cobra"
    "os"

    "github.com/intelligide/off-api-proxy/internal/app"
	_ "github.com/intelligide/off-api-proxy/internal/router"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Long:  ``,
	Run:   exec,
}

var (
    configFile string = "configs/config.toml"
)

func init() {
	rootCmd.AddCommand(runCmd)

    runCmd.Flags().StringVarP(&configFile, "config", "c", configFile, "Config file")
    runCmd.Flags().StringVar(&beego.BConfig.Listen.HTTPAddr, "address", beego.BConfig.Listen.HTTPAddr, "Address")
	runCmd.Flags().IntVarP(&beego.BConfig.Listen.HTTPPort, "port", "p", beego.BConfig.Listen.HTTPPort, "Port")
	runCmd.Flags().BoolVarP(&app.Verbose, "verbose", "v", app.Verbose, "Verbose")
}

func exec(cmd *cobra.Command, args []string) {

    c, err := config.NewConfig("toml", configFile)
    if err != nil {
        beego.Emergency(err)
        os.Exit(1)
    }
    app.Config = app.AppConfig{BConfig: c}

    app.Init()

	if app.Verbose {
		beego.SetLevel(beego.LevelDebug)
	} else {
		beego.SetLevel(app.Config.LogLevel())
	}

	// config.GlobalConfig.ConfigureBeego(beego.BConfig)

	toolbox.StartTask()
	defer toolbox.StopTask()

	beego.Run()
}
