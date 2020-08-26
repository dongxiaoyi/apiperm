package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/pkg/errors"
	"github.com/titpetric/factory/resputil"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

// chi api 配置
type HttpConfig struct {
	Addr	string
}

// 默认用户的密码
type DefaultPassword struct {
	Admin string
	Public string
}

// neo4j 配置
type Neo4jConfig struct {
	Addr string
	UserName string
	Password string
}

// JWT配置
type JWTConfig struct {
	TokenClaim string
	TokenAuth  *jwtauth.JWTAuth
}

func (jwtcfg *JWTConfig) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(jwtcfg.TokenAuth)
}

func (jwtcfg *JWTConfig) Encoding(name string) string {
	claims := jwt.MapClaims{jwtcfg.TokenClaim: name}

	jwtauth.SetExpiryIn(claims, 36000 * time.Second)
	jwtauth.SetIssuedNow(claims)

	_, tokenString, _ := jwtcfg.TokenAuth.Encode(claims)
	return tokenString
}

func (jwtcfg *JWTConfig) Authenticate(r *http.Request) (string, string, error) {
	// 从cookie获取jwt，做验证
	userName := ""

	token := jwtauth.TokenFromCookie(r)
	if token == "" {
		return "", userName, errors.New("Empty or invalid JWT")
	}

	t, err := jwtcfg.TokenAuth.Decode(token)
	if err != nil {
		return "", userName, err
	}
	if !t.Valid {
		return "", userName, errors.New("Invalid JWT")
	}


	for k, v := range t.Claims.(jwt.MapClaims) {
		if k == "user_name" {
			userName = v.(string)
		}

	}

	return "success", userName, nil
}

func (jwtcfg *JWTConfig) Decode(r *http.Request) (string, string) {
	val, userName, _ := jwtcfg.Authenticate(r)
	return val, userName
}

// 授权中间件
func (jwtcfg *JWTConfig) Authenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, userName, err := jwtcfg.Authenticate(r)
			if err != nil {
				resputil.JSON(w, err)
				return
			}
			// 认证通过，username不是 "" ：context传入user_name,以便后续资源权限验证
			ctx := context.WithValue(r.Context(), "user_name", userName)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}