package share

import "fmt"

func ProgressPercent(positionUS, lengthUS int64) int {
	if lengthUS <= 0 || positionUS <= 0 {
		return 0
	}

	percent := int(positionUS * 100 / lengthUS)

	if percent > 100 {
		return 100
	}

	return percent
}

func ProgressLabel(positionUS, lengthUS int64) string {
	return fmt.Sprintf("%d/100%%", ProgressPercent(positionUS, lengthUS))
}
