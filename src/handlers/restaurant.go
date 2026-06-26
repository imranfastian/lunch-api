package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imranfastian/lunch-api/src/config"
)

// RestaurantToday is a restaurant with only the current day's menu items.
type RestaurantToday struct {
	ID      int               `json:"id"`
	Name    string            `json:"name"`
	City    string            `json:"city"`
	Address string            `json:"address"`
	Cuisine string            `json:"cuisine"`
	Rating  float64           `json:"rating"`
	Day     string            `json:"day"`
	Menu    []config.MenuItem `json:"menu"`
}

// RestaurantNearby is a restaurant with its computed distance from the SciLifeLab office.
type RestaurantNearby struct {
	config.Restaurant
	DistanceKm float64 `json:"distance_km"`
}

// filterByCity returns restaurants whose City matches the given value (case-insensitive).
func filterByCity(rs []config.Restaurant, city string) []config.Restaurant {
	var out []config.Restaurant
	for _, r := range rs {
		if strings.EqualFold(r.City, city) {
			out = append(out, r)
		}
	}
	return out
}

// GetRestaurants godoc
// @Summary      List restaurants
// @Description  Returns all restaurants. Filter by city with ?city=stockholm or ?city=uppsala.
// @Produce      json
// @Param        city  query  string  false  "Filter by city (stockholm or uppsala)"
// @Success      200   {array}   config.Restaurant
// @Router       /api/restaurants [get]
func GetRestaurants(c *gin.Context) {
	rs := config.GetRestaurants()
	if city := c.Query("city"); city != "" {
		rs = filterByCity(rs, city)
	}
	c.JSON(http.StatusOK, rs)
}

// GetRestaurantsToday godoc
// @Summary      List restaurants with today's menu
// @Description  Returns all restaurants showing only the current weekday's menu items. Filter by city with ?city=.
// @Produce      json
// @Param        city  query  string  false  "Filter by city (stockholm or uppsala)"
// @Success      200   {array}   handlers.RestaurantToday
// @Router       /api/restaurants/today [get]
func GetRestaurantsToday(c *gin.Context) {
	rs := config.GetRestaurants()
	if city := c.Query("city"); city != "" {
		rs = filterByCity(rs, city)
	}
	day := time.Now().Weekday().String()
	result := make([]RestaurantToday, 0, len(rs))
	for _, r := range rs {
		result = append(result, RestaurantToday{
			ID:      r.ID,
			Name:    r.Name,
			City:    r.City,
			Address: r.Address,
			Cuisine: r.Cuisine,
			Rating:  r.Rating,
			Day:     day,
			Menu:    r.WeeklyMenu[day],
		})
	}
	c.JSON(http.StatusOK, result)
}

// GetNearbyRestaurants godoc
// @Summary      List restaurants near the SciLifeLab office
// @Description  Returns restaurants within the given radius (km) of the SciLifeLab office in the specified city, sorted by distance. Default radius is 5 km.
// @Produce      json
// @Param        city    query  string   true   "City: stockholm or uppsala"
// @Param        radius  query  number   false  "Search radius in kilometres (default 5)"
// @Success      200     {array}   handlers.RestaurantNearby
// @Failure      400     {object}  map[string]string
// @Router       /api/restaurants/nearby [get]
func GetNearbyRestaurants(c *gin.Context) {
	city := strings.ToLower(c.Query("city"))
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "city parameter is required (stockholm or uppsala)"})
		return
	}
	office, ok := officeCoords[city]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown city; use stockholm or uppsala"})
		return
	}

	radius := 5.0
	if raw := c.Query("radius"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 {
			radius = v
		}
	}

	var result []RestaurantNearby
	for _, r := range config.GetRestaurants() {
		if !strings.EqualFold(r.City, city) {
			continue
		}
		coords, known := restaurantCoords[r.ID]
		if !known {
			continue
		}
		d := haversineKm(office[0], office[1], coords[0], coords[1])
		if d <= radius {
			result = append(result, RestaurantNearby{Restaurant: r, DistanceKm: roundKm(d)})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].DistanceKm < result[j].DistanceKm
	})
	if result == nil {
		result = []RestaurantNearby{}
	}
	c.JSON(http.StatusOK, result)
}

// GetRestaurant godoc
// @Summary      Get a restaurant by ID
// @Produce      json
// @Param        id   path  int  true  "Restaurant ID"
// @Success      200  {object}  config.Restaurant
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/restaurants/{id} [get]
func GetRestaurant(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	for _, r := range config.GetRestaurants() {
		if r.ID == id {
			c.JSON(http.StatusOK, r)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// GetRestaurantMenu godoc
// @Summary      Get full weekly menu for a restaurant
// @Produce      json
// @Param        id   path  int  true  "Restaurant ID"
// @Success      200  {object}  config.WeeklyMenu
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/restaurants/{id}/menu [get]
func GetRestaurantMenu(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	for _, r := range config.GetRestaurants() {
		if r.ID == id {
			c.JSON(http.StatusOK, r.WeeklyMenu)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// GetRestaurantMenuToday godoc
// @Summary      Get today's menu for a restaurant
// @Description  Returns only the current weekday's menu items for the specified restaurant.
// @Produce      json
// @Param        id   path  int  true  "Restaurant ID"
// @Success      200  {object}  handlers.RestaurantToday
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/restaurants/{id}/menu/today [get]
func GetRestaurantMenuToday(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	day := time.Now().Weekday().String()
	for _, r := range config.GetRestaurants() {
		if r.ID == id {
			c.JSON(http.StatusOK, RestaurantToday{
				ID:      r.ID,
				Name:    r.Name,
				City:    r.City,
				Address: r.Address,
				Cuisine: r.Cuisine,
				Rating:  r.Rating,
				Day:     day,
				Menu:    r.WeeklyMenu[day],
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
}

// roundKm rounds a distance to two decimal places for cleaner JSON output.
func roundKm(d float64) float64 {
	return float64(int(d*100+0.5)) / 100
}

