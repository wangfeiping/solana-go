package exporter

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	PREFIX = "etl"
)

// Metrics holds all Prometheus metrics for the Solana ETL
type Metrics struct {
	TransactionsTotal  *prometheus.CounterVec
	BlockHeight        *prometheus.GaugeVec
	SubscriptionsActive *prometheus.GaugeVec
	ConnectionStatus   *prometheus.GaugeVec
	LogsQueue          *prometheus.GaugeVec
}

// NewMetrics creates a new Metrics instance with all Prometheus metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		TransactionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_transactions_total", PREFIX),
				Help: "Total number of transactions processed",
			},
			[]string{"chain", "mint_account", "status", "operation"},
		),

		BlockHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_block_height", PREFIX),
				Help: "Latest block height from transactions",
			},
			[]string{"chain"},
		),

		SubscriptionsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_subscriptions_active", PREFIX),
				Help: "Number of active WebSocket subscriptions",
			},
			[]string{"chain"},
		),

		ConnectionStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_connection_status", PREFIX),
				Help: "WebSocket connection status (1=connected, 0=disconnected)",
			},
			[]string{"chain"},
		),

		LogsQueue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: fmt.Sprintf("%s_logs_queue", PREFIX),
				Help: "Current number of messages in the processing queue",
			},
			[]string{"chain", "size"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(metrics.TransactionsTotal)
	prometheus.MustRegister(metrics.BlockHeight)
	prometheus.MustRegister(metrics.SubscriptionsActive)
	prometheus.MustRegister(metrics.ConnectionStatus)
	prometheus.MustRegister(metrics.LogsQueue)

	return metrics
}

// RecordTransaction records a transaction metric
func (m *Metrics) RecordTransaction(chain, mintAccount, status, operation string) {
	m.TransactionsTotal.WithLabelValues(chain, mintAccount, status, operation).Inc()
}

// UpdateBlockHeight updates the block height metric
func (m *Metrics) UpdateBlockHeight(chain string, height float64) {
	m.BlockHeight.WithLabelValues(chain).Set(height)
}

// SetSubscriptionActive sets the subscription active status
func (m *Metrics) SetSubscriptionActive(chain string, active float64) {
	m.SubscriptionsActive.WithLabelValues(chain).Set(active)
}

// SetConnectionStatus sets the connection status
func (m *Metrics) SetConnectionStatus(chain string, status float64) {
	m.ConnectionStatus.WithLabelValues(chain).Set(status)
}

// UpdateLogsQueue updates the logs queue metric
func (m *Metrics) UpdateLogsQueue(chain, size string, count float64) {
	m.LogsQueue.WithLabelValues(chain, size).Set(count)
}
