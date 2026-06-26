package handlers

import "math"

// officeCoords maps lowercase city name to [lat, lon] of the SciLifeLab office.
var officeCoords = map[string][2]float64{
	"stockholm": {59.3508, 18.0181}, // TomtebodavÃ¤gen 23A, 171 65 Solna
	"uppsala":   {59.8437, 17.6350}, // Husargatan 3 (BMC), 751 23 Uppsala
}

// restaurantCoords maps each restaurant ID to approximate [lat, lon].
// Derived from the addresses in restaurants_menu.json; kept here so the
// provided data file remains unchanged.
//
// Distances from each SciLifeLab office (Haversine):
//   Stockholm restaurants vs TomtebodavÃ¤gen 23A, Solna (59.3508, 18.0181):
//     ID  1 Kungsgatan 12          ~3.7 km
//     ID  3 VÃ¤sterlÃ¥nggatan 42     ~4.7 km
//     ID  5 GÃ¶tgatan 48            ~5.5 km
//     ID  7 Nybrogatan 15          ~4.9 km
//     ID  9 Norr MÃ¤larstrand 64    ~2.6 km
//     ID 11 Odengatan 56           ~2.1 km
//     ID 13 DjurgÃ¥rdsslÃ¤tten 46    ~6.1 km
//     ID 15 Nytorget 4             ~5.6 km
//   Uppsala restaurants vs Husargatan 3 BMC (59.8437, 17.6350):
//     ID  2 SvartbÃ¤cksgatan 27     ~1.5 km
//     ID  4 Ã–stra Ã…gatan 31        ~1.9 km
//     ID  6 Skolgatan 45           ~0.9 km  â† within 1 km
//     ID  8 Sankt Eriks GrÃ¤nd 2    ~1.7 km
//     ID 10 Olof Palmes Plats 1    ~1.7 km
//     ID 12 Riddartorget 1         ~1.5 km
//     ID 14 Fyristorg 6            ~1.7 km
var restaurantCoords = map[int][2]float64{
	1:  {59.3359, 18.0637}, // Smak & Bistro, Kungsgatan 12, Stockholm          ~3.7 km
	2:  {59.8567, 17.6424}, // Uppsala TrÃ¤dgÃ¥rdscafÃ©, SvartbÃ¤cksgatan 27        ~1.5 km
	3:  {59.3238, 18.0710}, // Gamla Stans Valv, VÃ¤sterlÃ¥nggatan 42, Stockholm  ~4.7 km
	4:  {59.8591, 17.6505}, // Fyris Krog & Bar, Ã–stra Ã…gatan 31, Uppsala       ~1.9 km
	5:  {59.3150, 18.0693}, // SÃ¶dermalm Matstudio, GÃ¶tgatan 48, Stockholm      ~5.5 km
	6:  {59.8510, 17.6328}, // Botanika KÃ¶k, Skolgatan 45, Uppsala              ~0.9 km
	7:  {59.3372, 18.0829}, // Ã–stermalm Smakbar, Nybrogatan 15, Stockholm      ~4.9 km
	8:  {59.8574, 17.6451}, // Saluhallen Bistro, Sankt Eriks GrÃ¤nd 2, Uppsala  ~1.7 km
	9:  {59.3275, 18.0228}, // Kungsholmen Grill, Norr MÃ¤larstrand 64, Stockholm ~2.6 km
	10: {59.8577, 17.6454}, // Stationen Uppsala, Olof Palmes Plats 1           ~1.7 km
	11: {59.3451, 18.0462}, // Vasastan Vin & Bistro, Odengatan 56, Stockholm   ~2.1 km
	12: {59.8576, 17.6381}, // Villa Anna, Riddartorget 1, Uppsala              ~1.5 km
	13: {59.3333, 18.1079}, // DjurgÃ¥rden Terrassen, DjurgÃ¥rdsslÃ¤tten 46        ~6.1 km
	14: {59.8583, 17.6484}, // Hambergs Fisk, Fyristorg 6, Uppsala              ~1.7 km
	15: {59.3153, 18.0762}, // Nytorget Urban Deli, Nytorget 4, Stockholm       ~5.6 km
}

// haversineKm returns the great-circle distance in kilometres between two lat/lon points.
func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

