package model

import (
	"context"
	"github.com/hpcloud/tail"
)

// TailTask tail任务对象
type TailTask struct {
	TailFile *tail.Tail
	Topic    string
	Ctx      context.Context
	Cancel   context.CancelFunc
}

func NewTailTask(tailFile *tail.Tail, topic string) *TailTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &TailTask{
		TailFile: tailFile,
		Topic:    topic,
		Ctx:      ctx,
		Cancel:   cancel,
	}
}
