package reqip // import "wickborn.net/reqip"

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	// register a "generic" plugin, like a directive or middleware
	caddy.RegisterPlugin("reqip", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
	httpserver.RegisterDevDirective("reqip", "")
}

func setup(c *caddy.Controller) error {
	var basePath string
	for c.Next() {
		if !c.NextArg() {
			return c.ArgErr()
		}
		basePath = c.Val()
		if c.NextArg() {
			return c.ArgErr()
		}
	}

	if basePath == "" {
		basePath = "/"
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return Handler{
			BasePath: basePath,
			Next:     next,
		}
	})

	return nil
}

type Handler struct {
	Next     httpserver.Handler
	BasePath string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if !httpserver.Path(r.URL.Path).Matches(h.BasePath) {
		return h.Next.ServeHTTP(w, r)
	}

	if r.Method != "GET" {
		return http.StatusMethodNotAllowed, nil
	}

	addr, _, err := net.SplitHostPort(r.RemoteAddr)
	var IP net.IP
	if err == nil {
		IP = net.ParseIP(addr)
		if IP == nil {
			fmt.Errorf("cannot parse %q to an IP", addr)
		}
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Write([]byte(IP.String()))
	return http.StatusOK, nil
}
