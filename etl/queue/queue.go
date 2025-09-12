package queue

import (
	"log"

	"github.com/gagliardetto/solana-go/rpc/ws"
)

// MessageQueueItem represents a queued message
type MessageQueueItem struct {
	LogResult        *ws.LogResult
	MintAccount      string
	SubscriptionType string
}

// MessageQueue manages the message queue for Solana ETL
type MessageQueue struct {
	queue chan MessageQueueItem
	size  int
}

// NewMessageQueue creates a new message queue with specified size
func NewMessageQueue(size int) *MessageQueue {
	return &MessageQueue{
		queue: make(chan MessageQueueItem, size),
		size:  size,
	}
}

// Send sends a message to the queue (non-blocking)
func (mq *MessageQueue) Send(item MessageQueueItem) bool {
	select {
	case mq.queue <- item:
		return true
	default:
		return false
	}
}

// Receive receives a message from the queue (blocking)
func (mq *MessageQueue) Receive() MessageQueueItem {
	return <-mq.queue
}

// Size returns the queue size
func (mq *MessageQueue) Size() int {
	return mq.size
}

// Length returns the current number of items in the queue
func (mq *MessageQueue) Length() int {
	return len(mq.queue)
}

// Capacity returns the remaining capacity of the queue
func (mq *MessageQueue) Capacity() int {
	return cap(mq.queue) - len(mq.queue)
}

// Close closes the queue
func (mq *MessageQueue) Close() {
	close(mq.queue)
}

// ProcessQueue processes messages from the queue
func (mq *MessageQueue) ProcessQueue(processor func(MessageQueueItem)) {
	log.Printf("🔄 Message queue processor started with size: %d", mq.size)

	for item := range mq.queue {
		processor(item)
	}

	log.Printf("🔄 Message queue processor stopped")
}
