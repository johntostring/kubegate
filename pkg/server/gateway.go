package server

import (
	"context"
	"fmt"
	"github.com/johntostring/kubegate/pkg/server/handler"
	"github.com/johntostring/kubegate/pkg/server/types"
	"github.com/johntostring/kubegate/pkg/tunnel"
	"github.com/kataras/iris/v12"
	"github.com/prometheus/common/log"
	"time"
)

var (
	GATEWAY = &types.Gateway{
		ProxyStore: &types.InMemoryStore{},
	}
)

func StartGateway(gw *types.Gateway) {
	app := iris.New()

	iris.RegisterOnInterrupt(func() {
		timeout := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		app.Shutdown(ctx)
	})

	appConfig := iris.WithConfiguration(iris.Configuration{
		EnableOptimizations:              true,
		Charset:                          "UTF-8",
		DisableInterruptHandler:          true,
		DisablePathCorrection:            true,
		DisablePathCorrectionRedirection: true,
	})
	app.Configure(appConfig)

	if gw.TunnelServerEnabled {
		log.Warnf("Tunnel token: %s", gw.TunnelToken)
		tunnelServer := tunnel.NewTunnelServer(gw.TunnelToken)
		GATEWAY.TunnelServer = tunnelServer
		handler.RegisterTunnel(app, gw)
	}

	handler.RegisterClustersHandler(app, gw)
	app.Run(iris.TLS(fmt.Sprintf(":%d", gw.HttpsPort), gw.HttpsCertFile, gw.HttpsKeyFile))
}
