package Snowflake

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node
var onec sync.Once

func Init(workId int64) (err error) {
	node, err = snowflake.NewNode(workId)
	if err != nil {
		return
	}
	return
}

func GenerateId() int64 {
	onec.Do(func() {
		Init(1)
	})
	return node.Generate().Int64()
}
