package storage

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	GetDocumentState(ctx context.Context, docID string) (content string, version int, err error)
	SetDocumentState(ctx context.Context, docID, content string, version int) error
	ClearDocumentState(ctx context.Context, docID string) error
	PushOperation(ctx context.Context, docID string, opData []byte) error
	PopOperation(ctx context.Context, docID string) ([]byte, error)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) getDocKey(docID string) string {
	return "doc_session:" + docID
}

func (c *RedisCache) GetDocumentState(ctx context.Context, docID string) (string, int, error) {
	key := c.getDocKey(docID)
	data, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return "", 0, err
	}
	if len(data) == 0 {
		return "", 0, redis.Nil
	}

	content := data["content"]
	version, _ := strconv.Atoi(data["version"])

	return content, version, nil
}

func (c *RedisCache) SetDocumentState(ctx context.Context, docID, content string, version int) error {
	key := c.getDocKey(docID)
	return c.client.HSet(ctx, key, map[string]interface{}{
		"content": content,
		"version": version,
	}).Err()
}

func (c *RedisCache) ClearDocumentState(ctx context.Context, docID string) error {
	key := c.getDocKey(docID)
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) PushOperation(ctx context.Context, docID string, opData []byte) error {
	key := "doc_ops:" + docID
	return c.client.LPush(ctx, key, opData).Err()
}

func (c *RedisCache) PopOperation(ctx context.Context, docID string) ([]byte, error) {
	key := "doc_ops:" + docID
	return c.client.LPop(ctx, key).Bytes()
}