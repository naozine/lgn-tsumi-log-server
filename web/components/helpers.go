package components

import (
	"fmt"
	"strconv"
	"strings"
)

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
