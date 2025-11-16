package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	ctxUserID ctxKey = "userID"
	ctxRoles  ctxKey = "roles"
)

func MustJWTSecret() []byte {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		panic("JWT_SECRET not set")
	}
	return []byte(sec)
}

func NewTokenGenerator(secret []byte) func(sub string, roles []string) (string, error) {
	return func(sub string, roles []string) (string, error) {
		claims := jwt.MapClaims{
			"sub":   sub,
			"roles": roles,
			"iat":   time.Now().Unix(),
			"exp":   time.Now().Add(15 * time.Minute).Unix(), // short TTL
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return t.SignedString(secret)
	}
}

func AuthJWT(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			tokStr := strings.TrimPrefix(authz, "Bearer ")
			tok, err := jwt.Parse(tokStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return secret, nil
			})
			if err != nil || !tok.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "bad token", http.StatusUnauthorized)
				return
			}
			uid := fmt.Sprint(claims["sub"])
			var roles []string
			if rs, ok := claims["roles"].([]any); ok {
				for _, x := range rs { roles = append(roles, fmt.Sprint(x)) }
			} else if rs, ok := claims["roles"].([]string); ok {
				roles = rs
			}
			ctx := context.WithValue(r.Context(), ctxUserID, uid)
			ctx = context.WithValue(ctx, ctxRoles, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRoles(needed ...string) func(http.Handler) http.Handler {
	need := map[string]struct{}{}
	for _, n := range needed { need[n] = struct{}{} }

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(ctxRoles)
			get := func() []string {
				if v, ok := val.([]string); ok { return v }
				if vv, ok := val.([]any); ok {
					out := make([]string, 0, len(vv))
					for _, x := range vv { out = append(out, fmt.Sprint(x)) }
					return out
				}
				return nil
			}
			roles := get()
			for _, have := range roles {
				if _, ok := need[have]; ok {
					next.ServeHTTP(w, r); return
				}
			}
			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

func UserIDFromCtx(ctx context.Context) (string, error) {
	v := ctx.Value(ctxUserID)
	if s, ok := v.(string); ok && s != "" { return s, nil }
	return "", errors.New("no user in context")
}