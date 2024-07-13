package grpc

import (
	"context"
	tagv1 "github.com/dadaxiaoxiao/api-repository/api/proto/gen/tag/v1"
	"github.com/dadaxiaoxiao/tag/internal/domain"
	"github.com/dadaxiaoxiao/tag/internal/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type TagServiceServer struct {
	tagv1.UnimplementedTagServiceServer
	svc service.TagService
}

func NewTagServiceServer(svc service.TagService) *TagServiceServer {
	return &TagServiceServer{svc: svc}
}

func (t *TagServiceServer) Register(server *grpc.Server) {
	tagv1.RegisterTagServiceServer(server, t)
}

func (t *TagServiceServer) CreateTag(ctx context.Context, req *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	id, err := t.svc.CreateTag(ctx, req.GetUid(), req.GetName())
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:   id,
			Uid:  req.Uid,
			Name: req.Name,
		},
	}, err

}
func (t *TagServiceServer) GetTags(ctx context.Context, req *tagv1.GetTagsRequest) (*tagv1.GetTagResponse, error) {
	tags, err := t.svc.GetTags(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetTagResponse{
		Tag: slice.Map(tags, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}
func (t *TagServiceServer) AttachTags(ctx context.Context, req *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.svc.AttachTags(ctx, req.GetUid(), req.GetBiz(), req.GetUid(), req.GetTids())
	return &tagv1.AttachTagsResponse{}, err
}
func (t *TagServiceServer) GetBizTags(ctx context.Context, req *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	tags, err := t.svc.GetBizTags(ctx, req.GetUid(), req.GetBiz(), req.GetBizId())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetBizTagsResponse{
		Tags: slice.Map(tags, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
