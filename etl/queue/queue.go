package queue

import (
	"log"
	"strconv"

	"github.com/gagliardetto/solana-go/rpc/ws"
)

// MessageQueueItem represents a queued message
type MessageQueueItem struct {
	LogResult        *ws.LogResult
	MintAccount      string
	MintIndex        int
	SubscriptionType string
}

// MetricsUpdater interface for updating metrics
type MetricsUpdater interface {
	UpdateLogsQueue(chain, size string, count float64)
}

// MessageQueue manages the message queue for Solana ETL
type MessageQueue struct {
	queue   chan MessageQueueItem
	size    int
	metrics MetricsUpdater
	chain   string
}

// NewMessageQueue creates a new message queue with specified size
func NewMessageQueue(size int) *MessageQueue {
	return &MessageQueue{
		queue: make(chan MessageQueueItem, size),
		size:  size,
		chain: "solana", // Default chain
	}
}

// NewMessageQueueWithMetrics creates a new message queue with metrics support
func NewMessageQueueWithMetrics(size int, metrics MetricsUpdater, chain string) *MessageQueue {
	return &MessageQueue{
		queue:   make(chan MessageQueueItem, size),
		size:    size,
		metrics: metrics,
		chain:   chain,
	}
}

// Send sends a message to the queue (non-blocking)
func (mq *MessageQueue) Send(item MessageQueueItem) bool {
	select {
	case mq.queue <- item:
		// Update metrics after successful send
		if mq.metrics != nil {
			mq.metrics.UpdateLogsQueue(mq.chain, strconv.Itoa(mq.size), float64(len(mq.queue)))
		}
		return true
	default:
		return false
	}
}

// Receive receives a message from the queue (blocking)
func (mq *MessageQueue) Receive() MessageQueueItem {
	item := <-mq.queue
	// Update metrics after receiving
	if mq.metrics != nil {
		mq.metrics.UpdateLogsQueue(mq.chain, strconv.Itoa(mq.size), float64(len(mq.queue)))
	}
	return item
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
		// Update metrics before processing
		if mq.metrics != nil {
			mq.metrics.UpdateLogsQueue(mq.chain, strconv.Itoa(mq.size), float64(len(mq.queue)))
		}
		
		processor(item)
		
		// Update metrics after processing
		if mq.metrics != nil {
			mq.metrics.UpdateLogsQueue(mq.chain, strconv.Itoa(mq.size), float64(len(mq.queue)))
		}
	}
	
	log.Printf("🔄 Message queue processor stopped")
}
