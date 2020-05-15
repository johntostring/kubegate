package types

import (
	"crypto/tls"
	"github.com/johntostring/kubegate/pkg/tunnel"
)

type Gateway struct {
	TunnelServer        *tunnel.Server
	ProxyStore          Store
	JwtPublicKey        []byte
	HttpsPort           int                 `mapstructure:"https-port"`
	HttpsCertFile       string              `mapstructure:"https-certfile"`
	HttpsKeyFile        string              `mapstructure:"https-keyfile"`
	TunnelServerEnabled bool                `mapstructure:"tunnel-server-enabled"`
	TunnelClientEnabled bool                `mapstructure:"tunnel-client-enabled"`
	JwtPublicKeyFile    string              `mapstructure:"jwt-public-keyfile"`
	TunnelServerUrl     string              `mapstructure:"tunnel-server-url"`
	TunnelToken         string              `mapstructure:"tunnel-token"`
	Proxy               map[string]ProxySet `mapstructure:"proxy"`
}

type ProxySet struct {
	Url       string     `mapstructure:"url"`
	ClientTls *ClientTls `mapstructure:"clientTls"`
}
type ClientTls struct {
	CertFilePath string `mapstructure:"cert-file-path"`
	KeyFilePath  string `mapstructure:"key-file-path"`
	certificate  *tls.Certificate
}

func (clientTls *ClientTls) Certificate() (*tls.Certificate, error) {
	var (
		cert tls.Certificate
		err  error
	)
	if clientTls.certificate == nil {
		cert, err = tls.LoadX509KeyPair(clientTls.CertFilePath, clientTls.KeyFilePath)
		clientTls.certificate = &cert
	}
	return clientTls.certificate, err
}
