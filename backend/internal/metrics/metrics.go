package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Business Metrics
	MessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_messages_total",
			Help: "Total number of messages sent",
		},
		[]string{"room", "type"}, // type = "user" ή "gemini"
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "chat_active_connections",
			Help: "Number of active WebSocket connections",
		},
	)

	ActiveRooms = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "chat_active_rooms",
			Help: "Number of active chat rooms",
		},
	)

	AIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_ai_requests_total",
			Help: "Total number of Gemini AI requests",
		},
		[]string{"status"}, // status = "success" or "error"
	)

	DocumentsUploaded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_documents_uploaded_total",
			Help: "Total number of PDF documents uploaded",
		},
		[]string{"room"},
	)

	// Technical Metrics
	GeminiLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "chat_gemini_latency_seconds",
			Help:    "Gemini API response latency in seconds",
			Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30},
		},
	)

	WebSocketErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_websocket_errors_total",
			Help: "Total number of WebSocket errors",
		},
		[]string{"type"},
	)

	RedisPublishErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "chat_redis_publish_errors_total",
			Help: "Total number of Redis publish errors",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		MessagesTotal,
		ActiveConnections,
		ActiveRooms,
		AIRequestsTotal,
		DocumentsUploaded,
		GeminiLatency,
		WebSocketErrors,
		RedisPublishErrors,
	)
}
