// Copyright 2020 Fabian Wickborn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reqip // import "wickborn.net/reqip"

import (
	"fmt"
	"net"
	"net/http"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"
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
	if r.URL.Path != h.BasePath {
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
