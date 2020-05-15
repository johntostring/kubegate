package cmd

import (
	"github.com/johntostring/kubegate/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run the gateway",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		gw := server.GATEWAY
		err := viper.Unmarshal(&gw)
		if err != nil {
			panic(err)
		}
		if gw.HttpsPort == 0 {
			gw.HttpsPort = 8443
		}
		if gw.JwtPublicKeyFile != "" {
			gw.JwtPublicKey, err = ioutil.ReadFile(gw.JwtPublicKeyFile)
			if err != nil {
				panic(err)
			}
		}
		server.StartGateway(gw)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
