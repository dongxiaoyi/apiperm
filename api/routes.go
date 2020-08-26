package api

import (
	"apiring/api/inventory"
	"apiring/api/login"
	"apiring/api/logout"
	"apiring/api/profile"
	"apiring/configs"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"net/http"
)
var tokenAuth *jwtauth.JWTAuth


func ProtectedRouter() http.Handler {
	r := chi.NewRouter()

	// 加载中间件
	r.Use(configs.JWT.Authenticator())  // 接受验证的接口(线上需要打开)

	// 挂载子路由

	// 主路由
	// graphql接口
	r.Handle("/profile", handler.Playground("GraphQL playground protected", "/query"))
	r.Handle("/profile/query", handler.GraphQL(profile.NewExecutableSchema(profile.Config{Resolvers: &profile.Resolver{}})))

	r.Handle("/inventory", handler.Playground("GraphQL playground protected", "/inventory"))
	r.Handle("/inventory/query", handler.GraphQL(inventory.NewExecutableSchema(inventory.Config{Resolvers: &inventory.Resolver{}})))

	r.Post("/logout", logout.LogoutView)  // 登出接口为rest接口

	return r
}

func PublicRouter() http.Handler {
	r := chi.NewRouter()

	// 加载中间件

	// 挂载子路由

	// 主路由
	r.Post("/login", login.LoginView)  // 登录接口为rest接口，不需要通过jwt、权限验证
	//r.Post("/hangout", login.GohangoutView)  // gohangout测试的路由，可删

	return r
}

func MainRouter() http.Handler {
	r := chi.NewRouter()

	// 加载中间件
	r.Use(LoggerMiddleware()) // 日志中间件（必选项）
	cors := CorsMiddleware()
	r.Use(cors.Handler) // 跨域中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.DefaultCompress) // gzip压缩
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Recoverer)  // 从panic中恢复崩溃的服务
	r.Use(configs.JWT.Verifier())  // 每一个请求都可以收到jwt标识

	// 挂载子路由
	r.Mount("/protected", ProtectedRouter()) // 受保护的路由
	r.Mount("/public", PublicRouter()) // 如用户登录等接口不需要验证jwt

	return r
}