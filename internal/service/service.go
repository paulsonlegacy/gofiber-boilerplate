// Package service is the dependency hub for the application.
// Service wraps all infrastructure dependencies and is injected into usecases.
// Usecases never import infrastructure packages directly — they always go through Service.
// This keeps the domain layer clean and makes testing easier.
package service

import (
	"context"
	"fmt"

	"boilerplate/internal/config"
	"boilerplate/internal/domain/entity"
	"boilerplate/internal/infra/cache"
	"boilerplate/internal/infra/mail"
	"boilerplate/internal/infra/queue"
	"boilerplate/internal/infra/storage"
	"boilerplate/internal/infra/telegram"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service is the central dependency container. Inject it into usecases via their constructors.
// All infrastructure access from usecases goes through this struct.
type Service struct {
	Config   *config.Config
	DB       *gorm.DB
	Logger   *zap.Logger
	Mailer   *mail.MultiProviderMailer
	Queue    queue.Queue
	Storage  storage.Storage
	Telegram *telegram.TelegramService
	Cache    *cache.MemoryCache
	Context  context.Context

	Wallet *WalletService // low-level atomic wallet operations
}

// NewService constructs and wires the Service.
// Called once at startup — pass to all usecases via their constructors.
func NewService(
	ctx context.Context,
	cfg *config.Config,
	db *gorm.DB,
	logger *zap.Logger,
	mailer *mail.MultiProviderMailer,
	q queue.Queue,
	storage storage.Storage,
	telegram *telegram.TelegramService,
	cache *cache.MemoryCache,
) *Service {
	return &Service{
		Config:   cfg,
		DB:       db,
		Logger:   logger,
		Mailer:   mailer,
		Queue:    q,
		Storage:  storage,
		Telegram: telegram,
		Cache:    cache,
		Context:  ctx,
		Wallet:   NewWalletService(db),
	}
}

// StartBackgroundWorkers launches n goroutines that pull jobs from the queue and process them.
// Each worker type (mail, notification, etc.) is dispatched in the same loop via a type switch.
// Add new job types by adding a case to the switch in processJob.
func (s *Service) StartBackgroundWorkers(q queue.Queue, n int) {
	for i := 0; i < n; i++ {
		go func(workerID int) {
			s.Logger.Info("Worker started", zap.Int("worker_id", workerID))
			for job := range q.GetQueue() {
				s.processJob(job)
			}
			s.Logger.Info("Worker stopped", zap.Int("worker_id", workerID))
		}(i)
	}
	s.Logger.Info("Background workers started", zap.Int("count", n))
}

// processJob dispatches a job to the correct handler based on its type.
func (s *Service) processJob(job queue.Job) {
	switch job.Type {

	case queue.TypeMail:
		data, ok := job.Data.(queue.MailJobData)
		if !ok || data.Email == nil {
			return
		}
		if err := s.Mailer.Send(s.Context, *data.Email); err != nil {
			s.Logger.Error("Mail delivery failed", zap.Error(err), zap.Strings("to", data.Email.To))
		}

	case queue.TypeNotification:
		data, ok := job.Data.(queue.NotificationJobData)
		if !ok {
			return
		}
		notification := &entity.Notification{
			ActorID:   data.ActorID,
			ActorType: data.ActorType,
			Title:     data.Title,
			Message:   data.Message,
		}
		if err := s.DB.Create(notification).Error; err != nil {
			s.Logger.Error("Notification persistence failed", zap.Error(err))
		}

	default:
		s.Logger.Warn("Unknown job type received", zap.String("type", fmt.Sprintf("%s", job.Type)))
	}
}

// NotifySystem sends a Telegram alert. Use for critical system events that need human attention.
func (s *Service) NotifySystem(title, body string, isError bool) {
	if s.Telegram == nil {
		return
	}
	if isError {
		_ = s.Telegram.Error(title, body)
	} else {
		_ = s.Telegram.Info(title, body)
	}
}
