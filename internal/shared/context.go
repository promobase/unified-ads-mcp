package shared

import (
	"context"
)

// contextKey is a type for context keys
type contextKey string

const (
	// FacebookAccessTokenKey is the context key for Facebook access token
	FacebookAccessTokenKey contextKey = "facebook_access_token"
	// GoogleAccessTokenKey is the context key for Google access token
	GoogleAccessTokenKey contextKey = "google_access_token"
	// TikTokAccessTokenKey is the context key for TikTok access token
	TikTokAccessTokenKey contextKey = "tiktok_access_token"
	// EnabledObjectTypesKey is the context key for enabled object types
	EnabledObjectTypesKey contextKey = "enabled_object_types"
)

// FacebookAccessTokenFromContext retrieves the Facebook access token from context
func FacebookAccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(FacebookAccessTokenKey).(string)
	return token, ok
}

// WithFacebookAccessToken adds Facebook access token to context
func WithFacebookAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, FacebookAccessTokenKey, token)
}

// EnabledObjectTypesFromContext retrieves enabled object types from context
func EnabledObjectTypesFromContext(ctx context.Context) (map[string]bool, bool) {
	types, ok := ctx.Value(EnabledObjectTypesKey).(map[string]bool)
	return types, ok
}

// WithEnabledObjectTypes adds enabled object types to context
func WithEnabledObjectTypes(ctx context.Context, types map[string]bool) context.Context {
	return context.WithValue(ctx, EnabledObjectTypesKey, types)
}
