// @title SciLifeLab Lunch API
// @version 0.2.0
// @description REST API for SciLifeLab employees to explore lunch options at nearby restaurants.
// @host localhost:8000
// @BasePath /
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Registers the OpenAPI spec in the swag registry via init().
	// gin-swagger reads it from there for /swagger/doc.json â€” no static file needed.
	_ "github.com/imranfastian/lunch-api/docs"

	"github.com/imranfastian/lunch-api/src/config"
	"github.com/imranfastian/lunch-api/src/routes"
)

func main() {
	config.MustLoadConfig()    // load restaurants and users from JSON files in data/
	router := gin.Default()    // create Gin router with Logger and Recovery middleware
	routes.SetupRoutes(router) // register all API routes

	// Swagger UI at /swagger/index.html.
	// gin-swagger serves the spec from the swag registry at /swagger/doc.json.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

