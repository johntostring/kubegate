package cmd

import (
	"context"
	"fmt"
	"github.com/johntostring/kubegate/pkg/server"
	"github.com/johntostring/kubegate/pkg/server/handler"
	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"time"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		gateway := server.GATEWAY
		err := viper.Unmarshal(gateway)
		if err != nil {
			panic(err)
		}
		if gateway.JwtPublicKeyFile != "" {
			gateway.JwtPublicKey, err = ioutil.ReadFile(gateway.JwtPublicKeyFile)
			if err != nil {
				panic(err)
			}
		}

		go func() {
			server.StartProxy(gateway)
		}()

		//Create proxy server
		app := iris.New()
		iris.RegisterOnInterrupt(func() {
			timeout := 10 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			// close all hosts
			app.Shutdown(ctx)
		})

		config := iris.WithConfiguration(iris.Configuration{
			EnableOptimizations:              true,
			Charset:                          "UTF-8",
			DisableInterruptHandler:          true,
			DisablePathCorrection:            true,
			DisablePathCorrectionRedirection: true,
		})
		app.Configure(config)

		handler.RegisterClustersHandler(app, gateway)
		app.Run(iris.TLS(fmt.Sprintf(":%d", gateway.HttpsPort), gateway.HttpsCertFile, gateway.HttpsKeyFile))

	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// proxyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// proxyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
