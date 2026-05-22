package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	MessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_messages_total",
			Help: "Total messages sent",
		},
		[]string{"room", "type"}, // type = "user" ή "gemini"
	)

	ActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "chat_active_connections",
		Help: "Active WebSocket connections",
	})

	ActiveRooms = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "chat_active_rooms",
		Help: "Active chat rooms",
	})

	AIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_ai_requests_total",
			Help: "Total Gemini AI requests",
		},
		[]string{"status"},
	)

	DocumentsUploaded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_documents_uploaded_total",
			Help: "Total PDFs uploaded",
		},
		[]string{"room"},
	)

	GeminiLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "chat_gemini_latency_seconds",
		Help:    "Gemini API latency",
		Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30},
	})

	WebSocketErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_websocket_errors_total",
			Help: "WebSocket errors",
		},
		[]string{"type"},
	)

	RedisPublishErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_redis_publish_errors_total",
		Help: "Redis publish errors",
	})

	UserSignups = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_user_signups_total",
			Help: "Total user signups",
		},
		[]string{"status"},
	)

	UserLogins = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_user_logins_total",
			Help: "Total user logins",
		},
		[]string{"status"},
	)

	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chat_http_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	MongoErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_mongodb_errors_total",
			Help: "MongoDB errors",
		},
		[]string{"operation"},
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
		UserSignups,
		UserLogins,
		HTTPRequestsTotal,
		HTTPDuration,
		MongoErrors,
	)
}
