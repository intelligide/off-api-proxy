package commands

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/toolbox"
	"github.com/spf13/cobra"

	"github.com/intelligide/off-api-proxy/internal/app"
	_ "github.com/intelligide/off-api-proxy/internal/router"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "",
	Long:  ``,
	Run:   exec,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVar(&beego.BConfig.Listen.HTTPAddr, "address", beego.BConfig.Listen.HTTPAddr, "Address")
	runCmd.Flags().IntVarP(&beego.BConfig.Listen.HTTPPort, "port", "p", beego.BConfig.Listen.HTTPPort, "Port")
	runCmd.Flags().BoolVarP(&app.Verbose, "verbose", "v", app.Verbose, "Verbose")
}

func exec(cmd *cobra.Command, args []string) {
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
