package metrics

import "github.com/prometheus/client_golang/prometheus"

// UserMetrics contains business metrics for the Users service
type UserMetrics struct {
	UsersCreated    prometheus.Counter
	UsersDeleted    prometheus.Counter
	UsersUpdated    prometheus.Counter
	DBQueryDuration prometheus.Histogram
}

func NewUserMetrics(namespace string) *UserMetrics {
	m := &UserMetrics{
		UsersCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "users",
			Name:      "created_total",
			Help:      "Total number of users created",
		}),
		UsersDeleted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,	
			Subsystem: "users",
			Name:      "deleted_total",
			Help:      "Total number of users deleted",
		}),
		UsersUpdated: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "users",
			Name:      "updated_total",
			Help:      "Total number of users updated",
		}),
		DBQueryDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "database",
			Name:      "query_duration_seconds",
			Help:      "Duration of database queries in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0},
		}),
	}
	// Register all metrics
	prometheus.MustRegister(
		m.UsersCreated,
		m.UsersDeleted,
		m.UsersUpdated,
		m.DBQueryDuration,
	)

	return m
}
