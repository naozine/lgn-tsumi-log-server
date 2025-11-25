package appcontext

import (
	"context"
)

type contextKey string

const (
	userEmailKey  contextKey = "userEmail"
	isLoggedInKey contextKey = "isLoggedIn"
	hasPasskeyKey contextKey = "hasPasskey"
)

func WithUser(ctx context.Context, email string, loggedIn bool, hasPasskey bool) context.Context {
	ctx = context.WithValue(ctx, userEmailKey, email)
	ctx = context.WithValue(ctx, isLoggedInKey, loggedIn)
	ctx = context.WithValue(ctx, hasPasskeyKey, hasPasskey)
	return ctx
}

func GetUser(ctx context.Context) (string, bool, bool) {
	email, _ := ctx.Value(userEmailKey).(string)
	loggedIn, _ := ctx.Value(isLoggedInKey).(bool)
	hasPasskey, _ := ctx.Value(hasPasskeyKey).(bool)
	return email, loggedIn, hasPasskey
}
