package cc

import (
	"context"
	"net/http"
)

type clientCtxKey struct{}

func WithClient(ctx context.Context, client *http.Client) context.Context {
	return context.WithValue(ctx, clientCtxKey{}, client)
}

func Client(ctx context.Context) *http.Client {
	client, ok := ctx.Value(clientCtxKey{}).(*http.Client)
	if !ok || client == nil {
		return &http.Client{}
	}

	return client
}
