package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	ConnectionsTotal   prometheus.Gauge
	ConnectionsByRoom  *prometheus.GaugeVec
	MessagesReceived   prometheus.Counter
	MessagesSent       prometheus.Counter
	MessageLatency     prometheus.Histogram
	KafkaMessages      *prometheus.CounterVec
	RedisOperations    *prometheus.CounterVec
	AuthFailures       prometheus.Counter
}

func New() *Metrics {
	return &Metrics{
		ConnectionsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "ws_connections_total",
			Help: "Total number of active WebSocket connections",
		}),
		ConnectionsByRoom: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "ws_connections_by_room",
			Help: "Number of connections per room",
		}, []string{"room_type", "room_id"}),
		MessagesReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ws_messages_received_total",
			Help: "Total number of messages received from clients",
		}),
		MessagesSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ws_messages_sent_total",
			Help: "Total number of messages sent to clients",
		}),
		MessageLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "ws_message_latency_seconds",
			Help:    "Message processing latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		KafkaMessages: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "kafka_messages_processed_total",
			Help: "Total number of Kafka messages processed",
		}, []string{"topic", "status"}),
		RedisOperations: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		}, []string{"operation", "status"}),
		AuthFailures: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ws_auth_failures_total",
			Help: "Total number of authentication failures",
		}),
	}
}

func (m *Metrics) IncConnections() {
	m.ConnectionsTotal.Inc()
}

func (m *Metrics) DecConnections() {
	m.ConnectionsTotal.Dec()
}

func (m *Metrics) IncRoomConnections(roomType, roomID string) {
	m.ConnectionsByRoom.WithLabelValues(roomType, roomID).Inc()
}

func (m *Metrics) DecRoomConnections(roomType, roomID string) {
	m.ConnectionsByRoom.WithLabelValues(roomType, roomID).Dec()
}

func (m *Metrics) IncMessagesReceived() {
	m.MessagesReceived.Inc()
}

func (m *Metrics) IncMessagesSent() {
	m.MessagesSent.Inc()
}

func (m *Metrics) ObserveLatency(seconds float64) {
	m.MessageLatency.Observe(seconds)
}

func (m *Metrics) IncKafkaMessage(topic, status string) {
	m.KafkaMessages.WithLabelValues(topic, status).Inc()
}

func (m *Metrics) IncRedisOperation(operation, status string) {
	m.RedisOperations.WithLabelValues(operation, status).Inc()
}

func (m *Metrics) IncAuthFailures() {
	m.AuthFailures.Inc()
}
