package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dadaxiaoxiao/tag/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

// TagCache tag 缓存
type TagCache interface {
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	Append(ctx context.Context, uid int64, tag ...domain.Tag) error
	DelTag(ctx context.Context, uid int64) error
}

type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewRedisTagCache(client redis.Cmdable) TagCache {
	return &RedisTagCache{
		client:     client,
		expiration: time.Hour * 48,
	}
}

func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	key := r.userTagsKey(uid)
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(data))
	for _, ele := range data {
		var t domain.Tag
		err = json.Unmarshal([]byte(ele), &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	mapData := make(map[string]any)
	key := r.userTagsKey(uid)
	pip := r.client.Pipeline()
	for _, tag := range tags {
		// 转为json 字节流
		val, err := json.Marshal(tag)
		if err != nil {
			return err
		}
		mapData[strconv.FormatInt(tag.Id, 10)] = val
	}
	pip.HSet(ctx, key, mapData)
	// 无法辨别 key 是不是已经有过期时间，
	// 设置过期时间
	pip.Expire(ctx, key, r.expiration)
	_, err := pip.Exec(ctx)
	return err
}

func (r *RedisTagCache) DelTag(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.userTagsKey(uid)).Err()
}

func (r *RedisTagCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tag:%d", uid)
}