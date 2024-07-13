package service

import (
	"context"
	"github.com/dadaxiaoxiao/tag/internal/events"

	"github.com/dadaxiaoxiao/tag/internal/domain"
	"github.com/dadaxiaoxiao/tag/internal/repository"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type TagService interface {
	CreateTag(ctx context.Context, uid int64, name string) (int64, error)
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
}

type tagService struct {
	repo     repository.TagRepository
	producer events.Producer
}

func NewTagService(repo repository.TagRepository, producer events.Producer) TagService {
	return &tagService{
		repo:     repo,
		producer: producer,
	}
}

func (svc *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return svc.repo.CreateTag(ctx, domain.Tag{
		Uid:  uid,
		Name: name,
	})
}

func (svc *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return svc.repo.GetTags(ctx, uid)
}

func (svc *tagService) AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error {
	err := svc.repo.BindTagToBiz(ctx, uid, biz, bizId, tags)
	if err != nil {
		return err
	}

	// 异步接入搜索服务
	go func() {
		// 查询所有 tags
		tags, err := svc.repo.GetTagsById(ctx, tags)
		if err != nil {
			// 记录日志
			return
		}
		// 生产者发送消息
		// 这里这里是异步，可以单独使用超时控制，不需要控制在整个链路路由下
		pctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = svc.producer.ProducerSyncEvent(pctx, events.BizTags{
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
			Tags: slice.Map(tags, func(idx int, src domain.Tag) string {
				return src.Name
			}),
		})
		cancel()
		if err != nil {
			// 记录日志
		}
	}()
	return err

}

func (svc *tagService) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	return svc.repo.GetBizTags(ctx, uid, biz, bizId)
}
