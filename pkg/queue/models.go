package queue

import "sync"

var (
	queueMap      = make(map[string][]*chan Result)
	queueMapMutex = sync.RWMutex{}
)

type Result struct {
	Result string
	Cache  bool
	Error  error
}
