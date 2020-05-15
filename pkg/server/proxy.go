package server

import (
	"crypto/tls"
	"github.com/johntostring/kubegate/pkg/server/types"
	"github.com/johntostring/kubegate/pkg/tunnel"
)

func StartProxy(gateway *types.Gateway) {
	var err error
	upstreamMap := make(map[string]string)

	for name, p := range gateway.Proxy {
		if p.ClientTls.CertFilePath == "" || p.ClientTls.KeyFilePath == "" {
			err = gateway.ProxyStore.AddDirectProxy(name, p.Url)
		} else {
			var cert *tls.Certificate
			cert, err = p.ClientTls.Certificate()
			if err != nil {
				panic(err)
			} else {
				err = gateway.ProxyStore.AddClientTlsCertDirectProxy(name, p.Url, cert)
			}
		}
		upstreamMap[name] = p.Url
	}

	client := tunnel.Client{
		Remote:      gateway.TunnelServerUrl,
		UpstreamMap: upstreamMap,
		Token:       gateway.TunnelToken,
	}
	if err := client.Connect(); err != nil {
		panic(err)
	}
}
