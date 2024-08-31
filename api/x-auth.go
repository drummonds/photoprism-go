package api

import (
	"context"
	"net/http"
)

type ClientToken struct {
	token string
}

func NewXAuthProvider(token string) *ClientToken {
	return &ClientToken{token: token}
}

func (s *ClientToken) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("X-Auth-Token", s.token)
	// req.Header.Set("Authorization", "Bearer "+s.token)
	return nil
}
