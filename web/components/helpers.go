package components

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StopTiming は停車地の到着・出発時刻（動的計算結果）
type StopTiming struct {
	StopID        int64
	Arrived       bool       // 到着したか
	ArrivalTime   *time.Time // エリアに入った時刻
	DepartureTime *time.Time // エリアを出た時刻（nilなら滞在中）
}

// ArrivalTimeStr は到着時刻を "HH:MM" 形式で返す
func (t *StopTiming) ArrivalTimeStr() string {
	if t.ArrivalTime == nil {
		return ""
	}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return t.ArrivalTime.In(jst).Format("15:04")
}

// DepartureTimeStr は出発時刻を "HH:MM" 形式で返す
func (t *StopTiming) DepartureTimeStr() string {
	if t.DepartureTime == nil {
		return ""
	}
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return t.DepartureTime.In(jst).Format("15:04")
}

// StayDurationStr は滞在時間を "X分" 形式で返す
func (t *StopTiming) StayDurationStr() string {
	if t.ArrivalTime == nil || t.DepartureTime == nil {
		return "-"
	}
	duration := t.DepartureTime.Sub(*t.ArrivalTime)
	return fmt.Sprintf("%d分", int(duration.Minutes()))
}

// CalculateStayDuration は到着時刻と出発時刻の差分（分）を計算します
// timeStr1: 到着時刻 (HH:MM)
// timeStr2: 出発時刻 (HH:MM)
// 戻り値: 分数文字列 (例: "15分"), エラー時は "-"
func CalculateStayDuration(arrivalTime, departureTime string) string {
	if arrivalTime == "" || departureTime == "" {
		return "-"
	}

	startMin, err1 := parseTimeToMinutes(arrivalTime)
	endMin, err2 := parseTimeToMinutes(departureTime)

	if err1 != nil || err2 != nil {
		return "-"
	}

	diff := endMin - startMin
	// 日をまたぐ場合（例: 23:50 -> 00:10）は24時間(1440分)を加算
	if diff < 0 {
		diff += 1440
	}

	return fmt.Sprintf("%d分", diff)
}

// parseTimeToMinutes は "HH:MM" 形式の時間を分数に変換します
func parseTimeToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format")
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return hours*60 + minutes, nil
}
