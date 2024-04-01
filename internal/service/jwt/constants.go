package jwt

type key string

var (
	ContextKey key = "ContextAuthToken"
	CookieKey      = "AuthToken"
)
