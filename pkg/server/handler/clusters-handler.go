package handler

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/johntostring/kubegate/pkg/server/types"
	"github.com/johntostring/kubegate/pkg/tunnel"
	"github.com/kataras/iris/v12"
	"github.com/twinj/uuid"
	"k8s.io/apimachinery/pkg/util/proxy"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func RegisterClustersHandler(app *iris.Application, gateway *types.Gateway) {

	app.Any("/clusters/{clusterName:string}/{remainingPath:path}", func(ctx iris.Context) {
		cluster := ctx.Params().GetString("clusterName")
		authProxy, found := gateway.ProxyStore.Lookup(cluster)
		if !found {
			ctx.StatusCode(http.StatusNotFound)
			return
		}

		if authProxy.ProxyType == types.DirectProxy {
			handleDirectProxy(ctx, *gateway, authProxy)
			return
		}

		if authProxy.ProxyType == types.TunnelProxy {
			route := gateway.TunnelServer.Lookup(cluster)
			if route == nil {
				ctx.StatusCode(http.StatusNotFound)
				return
			}
			handleTunnelProxy(ctx, *gateway, route)
		}
	})
}

func handleTunnelProxy(ctx iris.Context, gateway types.Gateway, route *tunnel.Route) {
	req, res := ctx.Request(), ctx.ResponseWriter()
	cluster := ctx.Params().GetString("clusterName")
	targetPath := fmt.Sprintf("/%s", ctx.Params().GetString("remainingPath"))

	tunnelId := uuid.Formatter(uuid.NewV4(), uuid.FormatHex)
	log.Printf("[Tunnel %s] access [cluster %s] %s %s %s", tunnelId, cluster, req.Host, req.Method, targetPath)
	req.Header.Set(tunnel.IdHeader, tunnelId)

	u := *req.URL
	u.Host = req.Host
	//u.Path = targetPath
	u.Scheme = route.Scheme

	httpProxy := proxy.NewUpgradeAwareHandler(&u, route.Transport, !gateway.TunnelServer.DisableWrapTransport, false, gateway.TunnelServer)
	httpProxy.ServeHTTP(res, req)
}

func handleDirectProxy(ctx iris.Context, gateway types.Gateway, authProxy *types.Proxy) {
	req, res := ctx.Request(), ctx.ResponseWriter()
	cluster := ctx.Params().GetString("clusterName")

	stringToken, err := ExtractToken(ctx)
	if err != nil {
		HandleUnauthorized(ctx, err)
		return
	}
	token, err := Validate(stringToken, gateway.JwtPublicKey)
	if err != nil {
		HandleUnauthorized(ctx, err)
		return
	}

	err = InjectAuthProxyHeaders(ctx, token)
	if err != nil {
		HandleUnauthorized(ctx, err)
		return
	}
	targetPath := fmt.Sprintf("/%s", ctx.Params().GetString("remainingPath"))
	//targetUrl := *req.URL
	//targetUrl.Path = targetPath
	//targetUrl.Host = authProxy.Host
	//targetUrl.Scheme = authProxy.Schema
	targetUrl := &url.URL{
		Scheme: authProxy.Schema,
		Host:   authProxy.Host,
		Path:   targetPath,
	}
	log.Printf("[Direct Proxy] access cluster(%s) %s %s %s", cluster, authProxy.Host, req.Method, targetPath)
	httpProxy := proxy.NewUpgradeAwareHandler(targetUrl, authProxy.Transport, false, false, authProxy)
	httpProxy.ServeHTTP(res, req)
}

func InjectAuthProxyHeaders(ctx iris.Context, token *jwt.Token) error {
	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid payload")
	}

	req := ctx.Request()
	for header := range req.Header {
		if strings.HasPrefix(header, "X-Remote-") {
			req.Header.Del(header)
		}
	}

	username, ok := payload["user_name"].(string)
	if ok && username != "" {
		req.Header.Set("X-Remote-User", username)
	}
	return nil
}

func ExtractToken(ctx iris.Context) (string, error) {
	jwtHeader := strings.Split(ctx.GetHeader("Authorization"), " ")
	if jwtHeader[0] == "Bearer" && len(jwtHeader) == 2 {
		return jwtHeader[1], nil
	}
	jwtQuery := ctx.Request().URL.Query().Get("token")
	if jwtQuery != "" {
		return jwtQuery, nil
	}
	return "", fmt.Errorf("no token found")
}

func Validate(uToken string, publicKey []byte) (*jwt.Token, error) {
	if len(uToken) == 0 {
		return nil, fmt.Errorf("token is empty")
	}

	token, err := jwt.Parse(uToken, func(token *jwt.Token) (i interface{}, err error) {
		return jwt.ParseRSAPublicKeyFromPEM(publicKey)
	})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		err := fmt.Errorf("invalid payload")
		log.Println(err)
		return nil, err
	}

	_, ok = payload["user_name"].(string)

	if !ok {
		err := fmt.Errorf("invalid payload")
		log.Println(err)
		return nil, err
	}

	return token, nil
}

func HandleUnauthorized(ctx iris.Context, err error) {
	message := fmt.Sprintf("Unauthorized, %v", err)
	ctx.ContentType("application/json")
	ctx.StatusCode(http.StatusUnauthorized)

	responseBody := fmt.Sprintf(`
	{
	  "kind": "Status",
	  "apiVersion": "v1",
	  "metadata": {},
	  "status": "Failure",
	  "message": "User Unauthorized",
	  "reason": "%s",
	  "code": 401
	}`, message)

	ctx.Text(responseBody)
}
