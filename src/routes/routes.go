package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/imranfastian/lunch-api/src/handlers"
)

// SetupRoutes registers all API routes on the given engine.
// Static path segments (today, nearby) are registered before the :id wildcard
// so Gin's radix-tree router resolves them with higher priority.
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET("/restaurants", handlers.GetRestaurants)
		api.GET("/restaurants/today", handlers.GetRestaurantsToday)
		api.GET("/restaurants/nearby", handlers.GetNearbyRestaurants)
		api.GET("/restaurants/:id", handlers.GetRestaurant)
		api.GET("/restaurants/:id/menu", handlers.GetRestaurantMenu)
		api.GET("/restaurants/:id/menu/today", handlers.GetRestaurantMenuToday)
	}
}

