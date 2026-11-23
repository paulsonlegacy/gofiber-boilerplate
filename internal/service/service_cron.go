// service_cron.go — periodic background jobs.
// All recurring/scheduled work belongs here.
// Register new cron jobs in StartCronJobs — do not scatter Start* calls in main.go.
package service

import (
	"context"
	"time"

	"boilerplate/internal/domain/entity"

	"go.uber.org/zap"
)

// Cron intervals — adjust here, not at call sites.
const (
	cacheCleanupInterval        = 5 * time.Minute
	refreshTokenCleanupInterval = 24 * time.Hour
)

// StartCronJobs is the single entry point for all periodic background jobs.
// Call once from main.go after the service is fully initialized.
func (s *Service) StartCronJobs() {
	s.StartCacheCleaner(cacheCleanupInterval)
	s.StartRefreshTokenCleaner()
	// Add new cron registrations here:
	// s.StartYourCustomCron()
}

// StartCacheCleaner runs a background goroutine that evicts expired cache entries.
func (s *Service) StartCacheCleaner(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			removed := s.Cache.Clear()
			if removed > 0 {
				s.Logger.Info("Cache cleanup completed", zap.Int("removed", removed))
			}
		}
	}()
	s.Logger.Info("Cache cleaner started", zap.Duration("interval", interval))
}

// StartRefreshTokenCleaner periodically deletes expired refresh tokens from the DB.
func (s *Service) StartRefreshTokenCleaner() {
	go func() {
		ticker := time.NewTicker(refreshTokenCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanExpiredRefreshTokens()
			case <-s.Context.Done():
				s.Logger.Info("Refresh token cleaner stopped")
				return
			}
		}
	}()
	s.Logger.Info("Refresh token cleaner started", zap.Duration("interval", refreshTokenCleanupInterval))
}

func (s *Service) cleanExpiredRefreshTokens() {
	result := s.DB.WithContext(context.Background()).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.RefreshToken{})

	if result.Error != nil {
		s.Logger.Error("Refresh token cleanup failed", zap.Error(result.Error))
		return
	}
	if result.RowsAffected > 0 {
		s.Logger.Info("Refresh token cleanup completed", zap.Int64("deleted", result.RowsAffected))
	}
}
