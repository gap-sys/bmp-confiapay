package handlers

import (
	"cobranca-bmp/cache"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

type WebhookController struct {
	logger          *slog.Logger
	location        *time.Location
	webhookService  WebhookService
	cobrancaService CobrancaService
	cache           *cache.RedisCache
}

func NewWebhookController(logger *slog.Logger, cache *cache.RedisCache,
	location *time.Location, webhookService WebhookService, cobrancaCreditoPessoalService CobrancaService,
) *WebhookController {
	return &WebhookController{
		location:        location,
		logger:          logger,
		webhookService:  webhookService,
		cobrancaService: cobrancaCreditoPessoalService,

		cache: cache,
	}
}

func (w *WebhookController) GetPrefix() string {
	return "/"
}

func (w *WebhookController) Route(r fiber.Router) {
	r.Post("/webhook", w.WebhookCobranca())
}
