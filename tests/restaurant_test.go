package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/imranfastian/lunch-api/src/config"
	"github.com/imranfastian/lunch-api/src/routes"
)

// TestMain ensures the config is loaded and the working directory is at repo root.
func TestMain(m *testing.M) {
	_ = os.Chdir("..")
	config.MustLoadConfig()
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	routes.SetupRoutes(r)
	return r
}

// --- GET /api/restaurants ---

func TestGetRestaurants(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants")
	assertStatus(t, w, http.StatusOK)

	var resp []config.Restaurant
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if len(resp) != 15 {
		t.Fatalf("expected 15 restaurants, got %d", len(resp))
	}
}

func TestGetRestaurantsFilterByCity(t *testing.T) {
	r := setupRouter()

	cases := []struct {
		city string
		want int
	}{
		{"stockholm", 8},
		{"Stockholm", 8}, // case-insensitive
		{"Uppsala", 7},
		{"uppsala", 7},
	}

	for _, tc := range cases {
		t.Run(tc.city, func(t *testing.T) {
			w := doGet(r, "/api/restaurants?city="+tc.city)
			assertStatus(t, w, http.StatusOK)

			var resp []config.Restaurant
			mustUnmarshal(t, w.Body.Bytes(), &resp)

			if len(resp) != tc.want {
				t.Fatalf("city=%q: expected %d restaurants, got %d", tc.city, tc.want, len(resp))
			}
			for _, restaurant := range resp {
				if !strings.EqualFold(restaurant.City, tc.city) {
					t.Errorf("city=%q: unexpected city %q in response", tc.city, restaurant.City)
				}
			}
		})
	}
}

// --- GET /api/restaurants/today ---

func TestGetRestaurantsToday(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/today")
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if len(resp) != 15 {
		t.Fatalf("expected 15 entries, got %d", len(resp))
	}

	expectedDay := time.Now().Weekday().String()
	for i, item := range resp {
		day, _ := item["day"].(string)
		if day != expectedDay {
			t.Errorf("item %d: expected day=%q, got %q", i, expectedDay, day)
		}
		if _, ok := item["menu"]; !ok {
			t.Errorf("item %d: missing 'menu' field", i)
		}
	}
}

func TestGetRestaurantsTodayFilterByCity(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/today?city=Uppsala")
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if len(resp) != 7 {
		t.Fatalf("expected 7 Uppsala entries, got %d", len(resp))
	}
}

// --- GET /api/restaurants/nearby ---

func TestGetNearbyRestaurantsStockholm(t *testing.T) {
	r := setupRouter()

	// 5 km from SciLifeLab Solna captures several Stockholm restaurants.
	w := doGet(r, "/api/restaurants/nearby?city=stockholm&radius=5")
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if len(resp) == 0 {
		t.Fatal("expected at least one Stockholm restaurant within 5 km")
	}
	for _, item := range resp {
		city, _ := item["city"].(string)
		if !strings.EqualFold(city, "stockholm") {
			t.Errorf("expected Stockholm restaurant, got city=%q", city)
		}
		if _, ok := item["distance_km"]; !ok {
			t.Error("expected 'distance_km' field in nearby response")
		}
	}
}

func TestGetNearbyRestaurantsUppsala(t *testing.T) {
	r := setupRouter()

	// All 7 Uppsala restaurants are within 3 km of the BMC office.
	w := doGet(r, "/api/restaurants/nearby?city=uppsala&radius=5")
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if len(resp) != 7 {
		t.Fatalf("expected all 7 Uppsala restaurants within 5 km, got %d", len(resp))
	}
}

func TestGetNearbyRestaurantsSortedByDistance(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/nearby?city=stockholm&radius=10")
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if len(resp) < 2 {
		t.Skip("not enough nearby restaurants to verify sort order")
	}
	for i := 1; i < len(resp); i++ {
		prev := resp[i-1]["distance_km"].(float64)
		curr := resp[i]["distance_km"].(float64)
		if curr < prev {
			t.Errorf("results not sorted: index %d (%.2f km) < index %d (%.2f km)", i, curr, i-1, prev)
		}
	}
}

// TestNearbyNeverMixesCities verifies that Stockholm results never contain Uppsala
// restaurants and vice versa, regardless of radius.
func TestNearbyNeverMixesCities(t *testing.T) {
	r := setupRouter()
	cases := []struct {
		searchCity string
		wrongCity  string
	}{
		{"stockholm", "Uppsala"},
		{"Uppsala", "stockholm"},
	}
	for _, tc := range cases {
		t.Run(tc.searchCity, func(t *testing.T) {
			w := doGet(r, "/api/restaurants/nearby?city="+tc.searchCity+"&radius=50")
			assertStatus(t, w, http.StatusOK)
			var resp []map[string]interface{}
			mustUnmarshal(t, w.Body.Bytes(), &resp)
			for _, item := range resp {
				city, _ := item["city"].(string)
				if strings.EqualFold(city, tc.wrongCity) {
					t.Errorf("city=%q: result contains restaurant from %q (id=%v name=%v)",
						tc.searchCity, city, item["id"], item["name"])
				}
			}
		})
	}
}

// TestNearbyUppsalaRadius1 verifies that Botanika KÃ¶k (~0.9 km from BMC) appears
// within a 1 km search and that no Stockholm restaurant ever slips in.
func TestNearbyUppsalaRadius1(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/nearby?city=Uppsala&radius=1")
	assertStatus(t, w, http.StatusOK)

	var resp []map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	for _, item := range resp {
		city, _ := item["city"].(string)
		if !strings.EqualFold(city, "Uppsala") {
			t.Errorf("radius=1 Uppsala search returned a non-Uppsala restaurant: %v (%v)", item["name"], city)
		}
		dist, _ := item["distance_km"].(float64)
		if dist > 1.0 {
			t.Errorf("restaurant %v is %.2f km away â€” outside the 1 km radius", item["name"], dist)
		}
	}
}

func TestGetNearbyRestaurantsMissingCity(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/nearby")
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetNearbyRestaurantsUnknownCity(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/nearby?city=gothenburg")
	assertStatus(t, w, http.StatusBadRequest)
}

// --- GET /api/restaurants/:id ---

func TestGetRestaurant(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/1")
	assertStatus(t, w, http.StatusOK)

	var resp config.Restaurant
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	if resp.ID != 1 {
		t.Fatalf("expected restaurant id 1, got %d", resp.ID)
	}
}

func TestGetRestaurantNotFound(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/9999")
	assertStatus(t, w, http.StatusNotFound)
}

func TestGetRestaurantInvalidID(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/abc")
	assertStatus(t, w, http.StatusBadRequest)
}

// --- GET /api/restaurants/:id/menu ---

func TestGetRestaurantMenu(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/1/menu")
	assertStatus(t, w, http.StatusOK)

	var menu map[string][]map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &menu)

	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, d := range days {
		if _, ok := menu[d]; !ok {
			t.Errorf("expected menu to contain day %q", d)
		}
	}
}

// --- GET /api/restaurants/:id/menu/today ---

func TestGetRestaurantMenuToday(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/1/menu/today")
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	mustUnmarshal(t, w.Body.Bytes(), &resp)

	expectedDay := time.Now().Weekday().String()
	day, _ := resp["day"].(string)
	if day != expectedDay {
		t.Errorf("expected day=%q, got %q", expectedDay, day)
	}
	if _, ok := resp["menu"]; !ok {
		t.Error("expected 'menu' field in today response")
	}
}

func TestGetRestaurantMenuTodayNotFound(t *testing.T) {
	r := setupRouter()

	w := doGet(r, "/api/restaurants/9999/menu/today")
	assertStatus(t, w, http.StatusNotFound)
}

// --- shared test helpers ---

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("expected status %d, got %d (body: %s)", want, w.Code, w.Body.String())
	}
}

func mustUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal failed: %v\nbody: %s", err, data)
	}
}

