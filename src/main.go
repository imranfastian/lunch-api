// @title SciLifeLab Lunch API
// @version 0.2.0
// @description REST API for SciLifeLab employees to explore lunch options at nearby restaurants.
// @host localhost:8000
// @BasePath /
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Registers the OpenAPI spec in the swag registry via init().
	// gin-swagger reads it from there for /swagger/doc.json - no static file needed.
	_ "github.com/imranfastian/lunch-api/docs"

	"github.com/imranfastian/lunch-api/src/config"
	"github.com/imranfastian/lunch-api/src/middleware"
	"github.com/imranfastian/lunch-api/src/routes"
)

func main() {
	config.MustLoadConfig() // load restaurants and users from JSON files in data/

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.StructuredLogger())
	router.Use(middleware.PrometheusMiddleware())

	routes.SetupRoutes(router)

	// Swagger UI at /swagger/index.html.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Metrics server on a separate port so Prometheus scrapes never pass through
	// the API middleware and inflate lunch_api_http_requests_total with scrape noise.
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9091"
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on :%s/metrics", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("SciLifeLab Lunch API listening on :%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
