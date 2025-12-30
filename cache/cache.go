package cache

import (
	"cobranca-bmp/config"
	"cobranca-bmp/helpers"
	"cobranca-bmp/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type PropostaService interface {
	UpdateAssync(data models.UpdateDBData) (bool, error)
}

type FileProducer interface {
	Produce(path string, data any) error
}

type RedisCache struct {
	appTerminated   bool
	client          *redis.Client
	ctx             context.Context
	loc             *time.Location
	logger          *slog.Logger
	dbCh            string
	sub             *redis.PubSub
	dbMsgs          <-chan *redis.Message
	fileProducer    FileProducer
	propostaService PropostaService
}

func NewRedis(url string, logger *slog.Logger, ctx context.Context, loc *time.Location, propostaService PropostaService, dbCH string, fileProducer FileProducer) (*RedisCache, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	opt.ReadTimeout = 0

	client := redis.NewClient(opt)
	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	var cache = &RedisCache{
		appTerminated:   false,
		client:          client,
		ctx:             ctx,
		logger:          logger,
		propostaService: propostaService,
		dbCh:            dbCH,
		loc:             loc,
		fileProducer:    fileProducer}

	return cache, nil

}

func (r *RedisCache) SetSubscriber() error {
	sub := r.client.Subscribe(r.ctx, r.dbCh)
	_, err := sub.Receive(r.ctx)
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao configurar pubsub", err.Error(), nil)
		return err
	}
	r.sub = sub
	r.dbMsgs = sub.Channel()
	return nil

}

func (r *RedisCache) SetToken(key, token string) error {

	exp := time.Now().Add(config.REDIS_TOKEN_EXP)
	expInFuse := exp.In(r.loc).String()

	var hashFields = []string{
		"token", token,
		"expiration", expInFuse,
	}

	err := r.client.HSet(r.ctx, key, hashFields).Err()
	r.client.ExpireAt(r.ctx, key, exp)
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao armazenar token em cache", err.Error(), token)
		return err
	}

	return nil

}

func (r *RedisCache) SetTokenExp(key, token, expiration string, exp time.Time) error {

	var hashFields = []string{
		"token", token,
		"expiration", expiration,
	}

	err := r.client.HSet(r.ctx, key, hashFields).Err()
	r.client.ExpireAt(r.ctx, key, exp)
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao armazenar token em cache", err.Error(), token)
		return err
	}

	return nil

}

func (r *RedisCache) GetToken(key string) (string, error) {

	token, err := r.client.HGet(r.ctx, key, "token").Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao obter token em cache", err.Error(), token)
		}
		return "", err
	}
	return token, nil

}

func (r *RedisCache) GetTokenWithExp(key string) (string, string, error) {

	token, err := r.client.HGet(r.ctx, key, "token").Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao obter token em cache", err.Error(), token)
		}
		return "", "", err
	}

	exp, err := r.client.HGet(r.ctx, key, "expiration").Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao obter token em cache", err.Error(), token)
		}
		return "", "", err
	}
	return token, exp, nil

}

func (r *RedisCache) GetError(message, field, context string) string {
	key := fmt.Sprintf("%s:%s", message, context)
	result := r.client.HGet(r.ctx, key, field).Val()
	if result == "" {
		key = fmt.Sprintf("%s:%s", context, context)
		result = r.client.HGet(r.ctx, key, field).Val()
	}
	return result
}

func (r *RedisCache) SaveData(key, field string, data any, exp ...time.Time) error {
	err := r.client.HSet(r.ctx, key, field, data).Err()
	if err != nil {

		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao salvar hash", err, data)
		return err
	}
	if len(exp) == 1 {
		err := r.client.ExpireAt(r.ctx, key, exp[0]).Err()

		if err != nil {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao setar expiração para o hash", err.Error(), data)
			return err
		}
	}
	return nil

}

func (r *RedisCache) GetData(key, field string, data any) error {
	return r.client.HGet(r.ctx, key, field).Scan(data)
}

func (r *RedisCache) Get(key string) (string, error) {
	query := r.client.Get(r.ctx, key)
	err := query.Err()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao obter valor de chave", err.Error(), key)

		}
		return "", err
	}
	result := query.Val()

	return result, err

}

func (r *RedisCache) Set(key string, val any, exp time.Duration) error {
	err := r.client.Set(r.ctx, key, val, exp).Err()
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao armazenar chave", err.Error(), key)
		return err
	}
	return nil

}

func (r *RedisCache) Del(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao armazenar chave", err.Error(), key)
		return err
	}
	return nil

}

func (r *RedisCache) GenID() int {
	/*//Gera um  id pelo Redis. Este número formará a idempotencyKey com o id das simulações
	id, err := r.client.Incr(r.ctx, "IdempotencyKey").Result()
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao obter ID único", err.Error(), nil)
		newID := 1
		r.client.Set(r.ctx, "IdempotencyKey", newID+1, 0)
		return newID
	}
	return int(id)*/
	return int(time.Now().In(r.loc).UnixNano())

}

func (r *RedisCache) GetTTL(key string) time.Duration {
	//Gera um  id pelo Redis. Este número formará a idempotencyKey com o id das simulações
	ttl, err := r.client.TTL(r.ctx, key).Result()
	if err != nil {
		return -1
	}
	return ttl

}

func (r *RedisCache) Publish(payload models.UpdateDBData) error {
	payloadBytes, err := json.Marshal(payload)
	key := fmt.Sprintf("%d:%d", payload.IdProposta, payload.StatusId)
	if err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao serializar mensagem para publicar no canal", err.Error(), payload)
		return err
	}
	for try := range config.DB_UPDATE_MAX_RETRIES {
		err = r.client.Publish(r.ctx, r.dbCh, payloadBytes).Err()
		if err != nil {
			message := fmt.Sprintf("Erro ao publicar no canal, tentativa %d", try+1)
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", message, err.Error(), payload)

		} else {

			if err := r.SaveData(key, key, payload); err != nil {
				return err
			}
			return nil
		}
	}
	path := fmt.Sprintf("%s.json", key)
	r.fileProducer.Produce(path, payload)
	return err
}

func (r *RedisCache) Consume() {
	for msg := range r.dbMsgs {
		var payload models.UpdateDBData
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Erro ao desserializar payload", err.Error(), msg.Payload)
		}
		helpers.LogInfo(r.ctx, r.logger, r.loc, "redis", "", "Payload recebido no Redis", msg.Payload)

		noConn, err := r.propostaService.UpdateAssync(payload)
		if err != nil {
			if noConn {
				time.Sleep(config.DB_QUEUE_DELAY)
				r.Publish(payload)
			}
		}
		key := fmt.Sprintf("%d:%d", payload.IdProposta, payload.StatusId)
		err = r.client.Expire(r.ctx, key, 0).Err()
		if err != nil {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", key, "Erro ao deletar payload processado no Redis", err, msg.Payload)
			continue
		}

	}

}

func (r *RedisCache) Check() (map[string]any, error) {
	var info = make(map[string]any)
	pingResult := r.client.Ping(r.ctx)
	if pingResult.Val() != "PONG" {
		info["erro"] = pingResult.Err().Error()
	} else {
		info["status"] = "ok"
	}

	return info, nil

}

func (r *RedisCache) GetServiceName() string {
	return "redis"
}

func (r *RedisCache) ClosePubSub() error {
	if r.sub != nil {
		if err := r.sub.Close(); err != nil {
			helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Falha ao fechar pubsub", err.Error(), nil)
			return err
		}
	}
	return nil
}

func (r *RedisCache) Close() error {
	r.appTerminated = true
	if err := r.ClosePubSub(); err != nil {
		return err
	}

	helpers.LogInfo(r.ctx, r.logger, r.loc, "redis", "", "pubsub fechado", nil)

	if err := r.client.Close(); err != nil {
		helpers.LogError(r.ctx, r.logger, r.loc, "redis", "", "Falha ao fechar cliente", err.Error(), nil)
		return err

	}
	helpers.LogInfo(r.ctx, r.logger, r.loc, "redis", "", "cliente fechado", nil)
	return nil
}
