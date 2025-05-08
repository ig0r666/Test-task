package ratelimiter

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testtask/limiter/config"
	"testtask/limiter/core"
	"time"
)

type RateLimiter struct {
	interval time.Duration
	log      *slog.Logger
	mu       sync.Mutex
	cfg      config.Config
}

func New(ctx context.Context, log *slog.Logger, cfg config.Config, db core.RateLimiterDB) *RateLimiter {

	limiter := &RateLimiter{
		interval: cfg.RateLimit.UpdateInterval,
		log:      log,
		cfg:      cfg,
	}

	// В фоне запускаем периодическое пополнение токенов клиентов с заданным воеменным интервалом
	go limiter.UpdateTokensJob(ctx, limiter.interval, db)

	return limiter
}

func (rl *RateLimiter) AllowClientRequest(ctx context.Context, clientID string, db core.RateLimiterDB) (bool, error) {
	client, err := db.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, core.ErrClientNotFound) {
			newClient := core.Client{
				ClientID: clientID,
				Capacity: rl.cfg.RateLimit.Capacity,
				Tokens:   rl.cfg.RateLimit.Capacity - 1,
			}
			err = db.CreateClient(ctx, newClient)
			if err != nil {
				rl.log.Error("failed to create client", "error", err)
				return false, err
			}
			return true, err
		}
		return false, err
	}

	if client.Tokens <= 0 {
		rl.log.Debug("rate limit exceeded", "client_id", clientID)
		return false, nil
	}

	if err := db.UpdateClientToken(ctx, clientID, client.Tokens-1); err != nil {
		return false, err
	}

	return true, nil
}

func (rl *RateLimiter) UpdateTokensJob(ctx context.Context, interval time.Duration, db core.RateLimiterDB) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			db.UpdateAllTokens(ctx)
			rl.mu.Unlock()
		case <-ctx.Done():
			rl.log.Info("stop token update job")
			return
		}
	}
}
