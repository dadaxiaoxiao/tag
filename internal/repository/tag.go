package repository

import (
	"context"
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/dadaxiaoxiao/tag/internal/domain"
	"github.com/dadaxiaoxiao/tag/internal/repository/cache"
	"github.com/dadaxiaoxiao/tag/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	//GetBizTags 获取业务资源对应的标签
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
	// BindTagToBiz 给业务资源绑定标签
	BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error

	GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error)
}

type CacheTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	log   accesslog.Logger
}

func NewTagRepository(dao dao.TagDAO, cache cache.TagCache, log accesslog.Logger) *CacheTagRepository {
	return &CacheTagRepository{dao: dao, cache: cache, log: log}
}

// PreloadUserTags 缓存预加载
func (repo *CacheTagRepository) PreloadUserTags(ctx context.Context) error {
	offset := 0
	batch := 100
	/*
					通常做法
					1. 全部查询，根据uid 分组，整合数据
				    2. 循环写入 redis
					缺点：数据量大时查询慢，且消耗内存资源
			        这里做法
			        1.批量查询
			        2.使用redis 数组结构 list , hash 把分组的功能交给redis
					//   list rpush k1 v1 k2 v2    获取 lrange key start stop
					//   hash hset key field value 获取  hkeys key
		            这里key 就是 tag:user_tags:uid 命名如下 服务（表）:资源:资源id
	*/

	for {
		// 这里限制了每次查询时间
		dbCtc, cancel := context.WithTimeout(ctx, time.Second)
		// 查询所有tag
		tags, err := repo.dao.GetTags(dbCtc, offset, batch)
		cancel()
		// 查询的tag 属于不同的用户
		for _, tag := range tags {
			// 这里也可以使用超时控制
			rctx, cancel := context.WithTimeout(ctx, time.Second)
			err = repo.cache.Append(rctx, tag.Uid, repo.toDomain(tag))
			cancel()
			if err != nil {
				// 记录日志，你可以中断，你也可以继续
				continue
			}
		}
		if len(tags) < batch {
			return nil
		}
		offset += batch
	}
}

func (repo *CacheTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := repo.dao.CreateTag(ctx, repo.toEntity(tag))
	if err != nil {
		return 0, err
	}
	// 更新缓存
	err = repo.cache.Append(ctx, tag.Uid, tag)
	if err != nil {
		repo.log.Error("tag 更新缓存失败",
			accesslog.Int64("uid", tag.Uid),
			accesslog.Error(err))
	}
	return id, nil
}

func (repo *CacheTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	// 快路径
	res, err := repo.cache.GetTags(ctx, uid)
	if err == nil {
		return res, nil
	}
	// 慢路径
	tags, err := repo.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	res = slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	})
	// 更新缓存
	err = repo.cache.Append(ctx, uid, res...)
	if err != nil {
		repo.log.Error("tag 更新缓存失败",
			accesslog.Int64("uid", uid),
			accesslog.Error(err))
	}
	return res, nil
}

func (repo *CacheTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	res := slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	})
	return res, nil
}

func (repo *CacheTagRepository) BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	var tagBizs []dao.TagBiz
	tagBizs = slice.Map(tags, func(idx int, src int64) dao.TagBiz {
		return dao.TagBiz{
			BizId: bizId,
			Biz:   biz,
			Uid:   uid,
			Tid:   src,
		}
	})
	return repo.dao.CreateTagBiz(ctx, tagBizs)
}

func (repo *CacheTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}

func (repo *CacheTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (repo *CacheTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
