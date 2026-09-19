package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"shorebird-server/internal/config"
	"shorebird-server/internal/handler"
	"shorebird-server/internal/svc"
)

var configFile = flag.String("f", "etc/shorebird-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// Format errors into Shorebird ErrorResponse JSON format
	httpx.SetErrorHandler(func(err error) (int, any) {
		return http.StatusBadRequest, map[string]string{
			"code":    "bad_request",
			"message": err.Error(),
		}
	})

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// Dart http.Client omits Content-Type or sets text/plain when sending json.encode string.
	// Rewrite to application/json so go-zero correctly parses JSON request body.
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if strings.HasPrefix(ct, "text/plain") || ct == "" {
				r.Header.Set("Content-Type", "application/json; charset=utf-8")
			}
			next(w, r)
		}
	})

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting shorebird-api server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
