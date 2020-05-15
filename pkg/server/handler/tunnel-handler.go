package handler

import (
	"github.com/johntostring/kubegate/pkg/server/types"
	"github.com/johntostring/kubegate/pkg/tunnel"
	"github.com/kataras/iris/v12"
	"net/http"
	"strings"
)

func RegisterTunnel(app *iris.Application, gateway *types.Gateway) {

	app.Any("/tunnel", func(ctx iris.Context) {
		req, resp := ctx.Request(), ctx.ResponseWriter()

		upstreams := req.Header[http.CanonicalHeaderKey(tunnel.UpstreamHeader)]
		for _, upstream := range upstreams {
			parts := strings.SplitN(upstream, "=", 2)
			if len(parts) != 2 {
				continue
			}
			gateway.ProxyStore.AddTunnelProxy(parts[0], parts[1])
		}

		gateway.TunnelServer.ServeHTTP(resp, req)
		gateway.TunnelServer.RemoveRouter(req)
	})
}
