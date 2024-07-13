package dao

import "context"

// Tag 标签
type Tag struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Name  string `gorm:"type=varchar(4096)"`
	Uid   int64  `gorm:"index"`
	Ctime int64
	Utime int64
}

// TagBiz 对资源打标签
type TagBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"index:biz_type_id;column:biz_id"`
	Biz   string `gorm:"index:biz_type_id"`
	// 这里可以冗余字段
	// 创建人
	Uid int64 `gorm:"index"`
	// 标签id
	Tid   int64
	Tag   *Tag `gorm:"ForeignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	Ctime int64
	Utime int64
}

// TagDAO tag 数据访问对象
// 这里的命名应该偏向数据操作
type TagDAO interface {
	// CreateTag 创建标签
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	// CreateTagBiz 给业务资源打标签
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	// GetTagsByUid 根据uid 获取标签，这里是用户获取自己所有创建的标签
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	// GetTagsByBiz 获取业务资源被标识的标签，这里是用户获取针对某资源且自己打上的标签
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	// GetTagsById 根据标签id 获取标签
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
	//GetTags 分页查询
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
}
