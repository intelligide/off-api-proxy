package commands

import (
	"fmt"
	"github.com/intelligide/off-api-proxy/internal/build_info"
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "off-proxy",
}

func init() {
	rootCmd.Version = build_info.LongVersion
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
