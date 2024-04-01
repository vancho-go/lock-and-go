package middlewares

import "github.com/vancho-go/lock-and-go/pkg/logger"

type Middlewares struct {
	log *logger.Logger
}

func NewMiddlewares(log *logger.Logger) *Middlewares {
	return &Middlewares{
		log: log,
	}
}
