package handler

import (
	"time"

	"boilerplate/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SystemHandler provides health check and metrics endpoints.
// These are public routes — no auth required.
type SystemHandler struct {
	Service *service.Service
	DB      *gorm.DB
}

func NewSystemHandler(svc *service.Service, db *gorm.DB) *SystemHandler {
	return &SystemHandler{Service: svc, DB: db}
}

// Health handles GET /health
// Returns 200 with DB connectivity status. Use as the liveness probe target in Kubernetes.
func (h *SystemHandler) Health(c *fiber.Ctx) error {
	dbStatus := "ok"
	sqlDB, err := h.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "degraded"
	}

	return c.JSON(map[string]any{
		"status":    "ok",
		"db":        dbStatus,
		"timestamp": time.Now().UTC(),
	})
}

// Metrics handles GET /metrics
// Returns basic runtime stats. Extend with prometheus or custom metrics as needed.
func (h *SystemHandler) Metrics(c *fiber.Ctx) error {
	sqlDB, _ := h.DB.DB()

	var dbStats map[string]any
	if sqlDB != nil {
		stats := sqlDB.Stats()
		dbStats = map[string]any{
			"openConnections": stats.OpenConnections,
			"inUse":           stats.InUse,
			"idle":            stats.Idle,
		}
	}

	return c.JSON(map[string]any{
		"timestamp": time.Now().UTC(),
		"db":        dbStats,
	})
}
