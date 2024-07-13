package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

const tagIndexName = "tag_index"

// SaramaProducer sarama 实现生产者
type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(client sarama.Client) (*SaramaProducer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{
		producer: p,
	}, nil
}

// ProducerSyncEvent  发送同步事件
func (s *SaramaProducer) ProducerSyncEvent(ctx context.Context, tags BizTags) error {
	// 转为字节流
	data, _ := json.Marshal(tags)
	evt := SyncDataEvent{
		IndexName: tagIndexName,
		DocID:     fmt.Sprintf("%d_%s_%d", tags.Uid, tags.Biz, tags.BizId),
		// 转为字符串
		Data: string(data),
	}
	data, _ = json.Marshal(evt)
	_, _, err := s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: evt.Topic(),
		Value: sarama.ByteEncoder(data),
	})
	return err
}
