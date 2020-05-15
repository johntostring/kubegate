package tunnel

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rancher/remotedialer"
	"github.com/twinj/uuid"
)

func (c *Client) Connect() error {
	headers := http.Header{}
	headers.Set(IdHeader, uuid.Formatter(uuid.NewV4(), uuid.FormatHex))
	for k, v := range c.UpstreamMap {
		headers.Add(UpstreamHeader, fmt.Sprintf("%s=%s", k, v))
	}
	if c.Token != "" {
		headers.Add("Authorization", "Bearer "+c.Token)
	}

	url := c.Remote
	if !strings.HasPrefix(url, "ws") {
		url = "ws://" + url
	}
	for {
		remotedialer.ClientConnect(context.Background(), url+"/tunnel", headers, nil, AllAllow, nil)
	}
}

func AllAllow(network, address string) bool {
	return true
}
