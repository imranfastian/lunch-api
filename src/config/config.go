package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type MenuItem struct {
	Type string `json:"type"`
	Dish string `json:"dish"`
}

type WeeklyMenu map[string][]MenuItem

type Restaurant struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	City       string     `json:"city"`
	Address    string     `json:"address"`
	Cuisine    string     `json:"cuisine"`
	Rating     float64    `json:"rating"`
	WeeklyMenu WeeklyMenu `json:"weekly_menu"`
}

type Config struct {
	Restaurants []Restaurant
}

var C *Config

// MustLoadConfig loads JSON data files into memory and exits on error.
func MustLoadConfig() {
	cfg := &Config{}

	rpath := filepath.Join("src", "data", "restaurants_menu.json")
	rb, err := os.ReadFile(rpath)
	if err != nil {
		log.Fatalf("failed to read restaurants data: %v", err)
	}
	if err := json.Unmarshal(rb, &cfg.Restaurants); err != nil {
		log.Fatalf("failed to unmarshal restaurants data: %v", err)
	}

	C = cfg
	log.Printf("Loaded %d restaurants", len(C.Restaurants))
}

func GetRestaurants() []Restaurant {
	return C.Restaurants
}

