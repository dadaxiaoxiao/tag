package events

import "context"

// Producer 生成者
type Producer interface {
	// ProducerSyncEvent 数据同步
	ProducerSyncEvent(ctx context.Context, data BizTags) error
}

// BizTags 业务打标签数据
type BizTags struct {
	Uid   int64  `json:"uid"`
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
	// 标签名称
	Tags []string `json:"tags"`
}

// SyncDataEvent 通用的同步数据
//
// 假如说用于同步 tag
// IndexName = tag_index
// DocID = "uid_biz_bizId"
// Data = {"uid": 123, "biz":"xxx, "bizId": 12，"tags":"[xxx,xxx2]"} // 注意这里给的是 string([]byte)
type SyncDataEvent struct {
	IndexName string `json:"indexName"`
	DocID     string `json:"docID"`
	Data      string `json:"data"`
}

func (SyncDataEvent) Topic() string {
	return "search_sync_data"
}
