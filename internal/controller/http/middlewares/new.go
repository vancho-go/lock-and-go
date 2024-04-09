package middlewares

import "github.com/vancho-go/lock-and-go/pkg/logger"

// Middlewares представляет тип для мидлвари.
type Middlewares struct {
	log *logger.Logger
}

// NewMiddlewares конструктор Middlewares.
func NewMiddlewares(log *logger.Logger) *Middlewares {
	return &Middlewares{
		log: log,
	}
}
