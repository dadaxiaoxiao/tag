package dao

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type GORMTagDAO struct {
	db *gorm.DB
}

func NewGORMTagDAO(db *gorm.DB) TagDAO {
	return &GORMTagDAO{
		db: db,
	}
}

func (dao *GORMTagDAO) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.Ctime = now
	tag.Utime = now
	err := dao.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

func (dao *GORMTagDAO) CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error {
	if len(tagBiz) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for _, t := range tagBiz {
		t.Ctime = now
		t.Utime = now
	}
	first := tagBiz[0]
	// 这里完成覆盖式的操作
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 如果没有 tag_biz 没有uid ，想要完成删除
		// delete from tag_bizs where tid in (select distinct id form tags where uid = ? ) and biz =? and biz_id = ?
		err := tx.Model(&TagBiz{}).Delete("uid = ? AND biz =? AND biz_id =?", first.Uid, first.BizId, first.BizId).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBiz).Error
	})
}

func (dao *GORMTagDAO) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("uid = ?", uid).Find(&res).Error
	return res, err
}

func (dao *GORMTagDAO) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {


	// 互联网做法
	var tbs []TagBiz
	err := dao.db.WithContext(ctx).Where("biz = ? and biz_id = ? and uid =?", biz, bizId, uid).Find(&tbs).Error
	if err != nil {
		return nil, err
	}
	ids := slice.Map(tbs, func(idx int, src TagBiz) int64 {
		return src.Tid
	})
	var res []Tag
	err = dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, err

}

func (dao *GORMTagDAO) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

func (dao *GORMTagDAO) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}
