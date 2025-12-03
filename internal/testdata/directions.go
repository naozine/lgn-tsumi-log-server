package testdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// DirectionsClient は Google Directions API のクライアント
type DirectionsClient struct {
	APIKey string
}

// NewDirectionsClient は新しいクライアントを作成する
func NewDirectionsClient(apiKey string) *DirectionsClient {
	return &DirectionsClient{APIKey: apiKey}
}

// GetRoute は2点間のルートを取得する
func (c *DirectionsClient) GetRoute(origin, destination LatLng) (*RouteResult, error) {
	baseURL := "https://maps.googleapis.com/maps/api/directions/json"

	params := url.Values{}
	params.Set("origin", fmt.Sprintf("%f,%f", origin.Lat, origin.Lng))
	params.Set("destination", fmt.Sprintf("%f,%f", destination.Lat, destination.Lng))
	params.Set("mode", "driving")
	params.Set("key", c.APIKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var result directionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "OK" {
		return nil, fmt.Errorf("API returned status: %s", result.Status)
	}

	if len(result.Routes) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	route := result.Routes[0]
	if len(route.Legs) == 0 {
		return nil, fmt.Errorf("no legs found in route")
	}

	leg := route.Legs[0]

	// overview_polyline をデコード
	points := DecodePolyline(route.OverviewPolyline.Points)

	return &RouteResult{
		Points:         points,
		DistanceMeters: leg.Distance.Value,
		DurationSecs:   leg.Duration.Value,
	}, nil
}

// RouteResult はルート取得結果
type RouteResult struct {
	Points         []LatLng // 経路上のポイント
	DistanceMeters int      // 総距離（メートル）
	DurationSecs   int      // 予想所要時間（秒）
}

// Google Directions API レスポンス構造体
type directionsResponse struct {
	Status string `json:"status"`
	Routes []struct {
		OverviewPolyline struct {
			Points string `json:"points"`
		} `json:"overview_polyline"`
		Legs []struct {
			Distance struct {
				Value int `json:"value"`
			} `json:"distance"`
			Duration struct {
				Value int `json:"value"`
			} `json:"duration"`
		} `json:"legs"`
	} `json:"routes"`
}
