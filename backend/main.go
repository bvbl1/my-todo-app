package main

import (
	"net/http"
	"time"

	api "github.com/el10savio/TODO-Fullstack-App-Go-Gin-Postgres-React/backend/api"

	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 1.5, 2, 3},
		},
		[]string{"method", "path"},
	)

	inFlightRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "Number of in-flight HTTP requests",
		},
	)
)

// Initialize Prometheus metrics
func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(inFlightRequests)
}

// Middleware to track Prometheus metrics
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// Skip metrics for Prometheus endpoint itself
		if path == "/metrics" {
			c.Next()
			return
		}

		inFlightRequests.Inc()
		defer inFlightRequests.Dec()

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, http.StatusText(status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// Function called for index
func indexView(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")
	c.JSON(http.StatusOK, gin.H{"message": "TODO APP"})
}

// Setup Gin Routes
func SetupRoutes() *gin.Engine {
	// Use Gin as router
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	// Add Prometheus middleware
	router.Use(prometheusMiddleware())

	// Set route for index
	router.GET("/", indexView)

	// Set routes for API
	router.GET("/items", api.TodoItems)
	router.GET("/item/create/:item", api.CreateTodoItem)
	router.GET("/item/update/:id/:done", api.UpdateTodoItem)
	router.GET("/item/delete/:id", api.DeleteTodoItem)

	// Add Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}

// Main function
func main() {
	api.SetupPostgres()
	router := SetupRoutes()
	router.Run(":8081")
}
