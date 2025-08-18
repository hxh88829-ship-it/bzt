package server

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	stdhttp "net/http"
	"strings"
	v1 "valueguard/api/helloworld/v1"
	"valueguard/internal/conf"
	"valueguard/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				jwt.Server(func(token *jwtv5.Token) (interface{}, error) {
					return []byte("123456"), nil
				}),
			).Match(func(ctx context.Context, operation string) bool {
				fmt.Println("Operation:", operation) //找到接口对应程序内的路径
				return !strings.HasSuffix(operation, "/helloworld.v1.Greeter/BindWallet") &&
					!strings.HasSuffix(operation, "/helloworld.v1.Greeter/LoginWithWallet") && // 这里是设置不用密钥访问的接口
					!strings.HasSuffix(operation, "/helloworld.v1.Greeter/GetLoginMessage") &&
					!strings.HasSuffix(operation, "/helloworld.v1.Greeter/OpenOrder") &&
					!strings.HasSuffix(operation, "/helloworld.v1.Greeter/CloseOrder") &&
					!strings.HasSuffix(operation, "/helloworld.v1.Greeter/GetAirdrop")

			}).Build(),
		),
		http.Filter(CorsFilter),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}

func CorsFilter(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		// 设置 CORS 头部
		w.Header().Set("Access-Control-Allow-Origin", "*") // 如果要指定来源，改成 http://localhost:3000 这类
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// OPTIONS 请求拦截处理
		if r.Method == "OPTIONS" {
			w.WriteHeader(stdhttp.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
