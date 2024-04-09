package jwt

type key string

var (
	// ContextKey ключ, обозначающий именования JWT ключа в контексте.
	ContextKey key = "ContextAuthToken"
	// CookieKey ключ, обозначающий именования JWT куки.
	CookieKey = "AuthToken"
)
