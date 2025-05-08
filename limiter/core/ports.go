package core

import (
	"context"
	"time"
)

// Здесь целенаправлено разделил интерфейс DB на два интерфейса - для crud и для limiter'a
// с идеей, что если появится необходимость заменить реализацию бд для лимитера, нам не придется изменять методы, не относящиеся к работе лимитера

type RateLimiterDB interface {
	GetClient(context.Context, string) (Client, error)
	CreateClient(context.Context, Client) error
	UpdateClientToken(context.Context, string, int) error
	UpdateClientCapacity(context.Context, string, int) error
	UpdateAllTokens(context.Context) error
}

type CrudDB interface {
	GetClient(context.Context, string) (Client, error)
	GetAllClients(context.Context) ([]Client, error)
	CreateClient(context.Context, Client) error
	RemoveClient(context.Context, string) error
	UpdateClientCapacity(context.Context, string, int) error
}

type RateLimiter interface {
	AllowClientRequest(context.Context, string, RateLimiterDB) (bool, error)
	UpdateTokensJob(context.Context, time.Duration, RateLimiterDB)
}
