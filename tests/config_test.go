package handlers

import (
	"os"
	"testing"

	"github.com/imranfastian/lunch-api/src/config"
)

func TestConfigLoaded(t *testing.T) {
	if config.C == nil {
		t.Fatalf("expected config.C to be initialized")
	}
	if len(config.C.Restaurants) == 0 {
		t.Fatalf("expected at least one restaurant in config")
	}
	if _, err := os.Stat("src/data/restaurants_menu.json"); err != nil {
		t.Fatalf("restaurants_menu.json missing: %v", err)
	}
}

