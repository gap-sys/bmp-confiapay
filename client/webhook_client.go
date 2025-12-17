package client

import (
	"cobranca-bmp/auth"
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type WebhookClient struct {
	APIClient
}

func NewWebhookClient(ctx context.Context, loc *time.Location, logger *slog.Logger) *WebhookClient {
	return &WebhookClient{APIClient: NewApiClient(ctx, loc, nil, logger, "", "", "")}

}

func (w *WebhookClient) RequestToWebhook(data any, url string) error {
	req, err := w.newRequest(http.MethodPost, url, data)
	if err != nil {
		return err
	}
	key, err := auth.EncryptHash(config.WEBHOOK_HASH, config.WEBHOOK_KEY)
	if err != nil {
		helpers.LogError(w.ctx, w.logger, w.loc, "webhook client", "", "Erro ao criptografar apikey", err.Error(), data)
		return err
	}
	req.Header.Set("X-API-Key", key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var response json.RawMessage
		response, _ = io.ReadAll(resp.Body)
		err := fmt.Errorf("Erro ao requisitar para o webhook, status code %d", resp.StatusCode)
		helpers.LogError(w.ctx, w.logger, w.loc, "webhook client", resp.Status, err.Error(), err.Error(), map[string]any{
			"header":   req.Header,
			"body":     data,
			"response": string(response),
		})
		return err
	}

	return nil

}
