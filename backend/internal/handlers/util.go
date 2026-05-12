package handlers

import (
	"time"
)

var shanghaiTZ *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	shanghaiTZ = loc
}

func ShanghaiToday() time.Time {
	now := time.Now().In(shanghaiTZ)
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, shanghaiTZ)
}

// DaysUntilPlanned: plannedDay − today（天数，负表示相对今天已过期）。
func DaysUntilPlanned(today, planned time.Time) int {
	t0 := today.In(shanghaiTZ)
	y0, m0, d0 := t0.Date()
	t1 := planned.In(shanghaiTZ)
	y1, m1, d1 := t1.Date()
	st := time.Date(y0, m0, d0, 0, 0, 0, 0, shanghaiTZ)
	sp := time.Date(y1, m1, d1, 0, 0, 0, 0, shanghaiTZ)
	return int(sp.Sub(st).Hours() / 24)
}

// AlertLevel: none | orange | red — 应付到期前15天橙色、前7天与逾期红色；应收逾期未付红色，未到期沿用15/7天规则。
func AlertLevel(contractType string, planned time.Time, isPaid bool, today time.Time) string {
	if isPaid {
		return "none"
	}
	d := DaysUntilPlanned(today, planned)
	if contractType == "sales" {
		if d < 0 {
			return "red"
		}
	} else if d < 0 {
		return "red"
	}
	if d <= 7 {
		return "red"
	}
	if d <= 15 {
		return "orange"
	}
	return "none"
}

func firstDayOfMonth(t time.Time) time.Time {
	y, m, _ := t.In(shanghaiTZ).Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, shanghaiTZ)
}

func lastDayOfMonth(t time.Time) time.Time {
	fd := firstDayOfMonth(t)
	nm := fd.AddDate(0, 1, 0)
	return nm.Add(-24 * time.Hour)
}

func monthRangeFor(t time.Time) (start time.Time, end time.Time) {
	start = firstDayOfMonth(t)
	end = lastDayOfMonth(t)
	return start, end
}
