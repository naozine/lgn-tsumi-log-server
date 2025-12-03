package testdata

// DecodePolyline は Google Maps エンコード済み polyline をデコードして緯度経度のリストを返す
// 参考: https://developers.google.com/maps/documentation/utilities/polylinealgorithm
func DecodePolyline(encoded string) []LatLng {
	var points []LatLng
	index := 0
	lat := 0
	lng := 0

	for index < len(encoded) {
		// 緯度をデコード
		shift := 0
		result := 0
		for {
			b := int(encoded[index]) - 63
			index++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		if result&1 != 0 {
			lat += ^(result >> 1)
		} else {
			lat += result >> 1
		}

		// 経度をデコード
		shift = 0
		result = 0
		for {
			b := int(encoded[index]) - 63
			index++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		if result&1 != 0 {
			lng += ^(result >> 1)
		} else {
			lng += result >> 1
		}

		points = append(points, LatLng{
			Lat: float64(lat) / 1e5,
			Lng: float64(lng) / 1e5,
		})
	}

	return points
}

// LatLng は緯度経度を表す
type LatLng struct {
	Lat float64
	Lng float64
}
