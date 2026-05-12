package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type dashRaw struct {
	NodeID           uint       `gorm:"column:node_id"`
	ContractID       uint       `gorm:"column:contract_id"`
	NodeName         string     `gorm:"column:node_name"`
	Amount           float64    `gorm:"column:amount"`
	PlannedDate      time.Time  `gorm:"column:planned_date"`
	IsTriggered      bool       `gorm:"column:is_triggered"`
	IsPaid           bool       `gorm:"column:is_paid"`
	ContractTitle    string     `gorm:"column:title"`
	Counterparty     string     `gorm:"column:counterparty"`
	ContractTypeEnum string     `gorm:"column:contract_type"`
}

func (h *Routes) dashboardUpcoming(c *gin.Context) {
	today := ShanghaiToday()
	start, end := monthRangeFor(today)

	var rows []dashRaw
	err := h.DB.Raw(`
SELECT
  n.id AS node_id,
  n.contract_id,
  n.node_name,
  n.amount,
  n.planned_date,
  n.is_triggered,
  n.is_paid,
  c.title,
  c.counterparty,
  c.type::text AS contract_type
FROM payment_nodes n
JOIN contracts c ON c.id = n.contract_id
WHERE NOT n.is_paid
  AND n.planned_date >= ?
  AND n.planned_date <= ?
ORDER BY n.planned_date ASC, n.id ASC
`, start, end).Scan(&rows).Error
	if err != nil {
		log.Println("dashboardUpcoming", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		alert := AlertLevel(r.ContractTypeEnum, r.PlannedDate, r.IsPaid, today)
		out = append(out, gin.H{
			"nodeId":        r.NodeID,
			"contractId":    r.ContractID,
			"nodeName":      r.NodeName,
			"amount":        r.Amount,
			"plannedDate":   r.PlannedDate.Format("2006-01-02"),
			"isTriggered":   r.IsTriggered,
			"isPaid":        r.IsPaid,
			"contractTitle": r.ContractTitle,
			"counterparty":  r.Counterparty,
			"contractType":  r.ContractTypeEnum,
			"alertLevel":    alert,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"month":     today.Format("2006-01"),
		"today":     today.Format("2006-01-02"),
		"rangeFrom": start.Format("2006-01-02"),
		"rangeTo":   end.Format("2006-01-02"),
		"items":     out,
	})
}

func (h *Routes) reportSummary(c *gin.Context) {
	today := ShanghaiToday()
	y, _, _ := today.Date()
	yearStart := time.Date(y, 1, 1, 0, 0, 0, 0, shanghaiTZ)
	yearEnd := time.Date(y, 12, 31, 0, 0, 0, 0, shanghaiTZ)
	monthStart, monthEnd := monthRangeFor(today)

	var totals struct {
		TotalCount    int64 `gorm:"column:total_count"`
		ActiveCount   int64 `gorm:"column:active_count"`
		CompletedCt   int64 `gorm:"column:completed_count"`
		TerminatedCt  int64 `gorm:"column:terminated_count"`
	}
	if err := h.DB.Raw(`
SELECT
  COUNT(*) AS total_count,
  COUNT(*) FILTER (WHERE status = 'active') AS active_count,
  COUNT(*) FILTER (WHERE status = 'completed') AS completed_count,
  COUNT(*) FILTER (WHERE status = 'terminated') AS terminated_count
FROM contracts
`).Scan(&totals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "counts failed"})
		return
	}

	var monthExpenseDue float64
	h.DB.Raw(`
SELECT COALESCE(SUM(n.amount), 0)
FROM payment_nodes n
JOIN contracts c ON c.id = n.contract_id
WHERE NOT n.is_paid
  AND n.planned_date >= ? AND n.planned_date <= ?
  AND c.type <> 'sales'
`, monthStart, monthEnd).Scan(&monthExpenseDue)

	var monthExpensePaid float64
	h.DB.Raw(`
SELECT COALESCE(SUM(ap.amount), 0)
FROM actual_payments ap
JOIN contracts c ON c.id = ap.contract_id
WHERE ap.pay_date >= ? AND ap.pay_date <= ?
  AND c.type <> 'sales'
`, monthStart, monthEnd).Scan(&monthExpensePaid)

	var monthRecvDue float64
	h.DB.Raw(`
SELECT COALESCE(SUM(n.amount), 0)
FROM payment_nodes n
JOIN contracts c ON c.id = n.contract_id
WHERE NOT n.is_paid
  AND n.planned_date >= ? AND n.planned_date <= ?
  AND c.type = 'sales'
`, monthStart, monthEnd).Scan(&monthRecvDue)

	var monthRecvCollected float64
	h.DB.Raw(`
SELECT COALESCE(SUM(ap.amount), 0)
FROM actual_payments ap
JOIN contracts c ON c.id = ap.contract_id
WHERE ap.pay_date >= ? AND ap.pay_date <= ?
  AND c.type = 'sales'
`, monthStart, monthEnd).Scan(&monthRecvCollected)

	type typeRow struct {
		Type      string  `json:"type" gorm:"column:ctype"`
		AmountSum float64 `json:"amount" gorm:"column:amt"`
		CountCt   int64   `json:"count" gorm:"column:ct"`
	}
	var byType []typeRow
	if err := h.DB.Raw(`
SELECT type::text AS ctype, COALESCE(SUM(total_amount), 0)::float8 AS amt, COUNT(*)::bigint AS ct
FROM contracts GROUP BY type ORDER BY type
`).Scan(&byType).Error; err != nil {
		log.Println("byType", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "distribution failed"})
		return
	}

	var yearIncome, yearExpense float64
	h.DB.Raw(`
SELECT COALESCE(SUM(ap.amount), 0)
FROM actual_payments ap
JOIN contracts c ON c.id = ap.contract_id
WHERE c.type = 'sales' AND ap.pay_date >= ? AND ap.pay_date <= ?
`, yearStart, yearEnd).Scan(&yearIncome)
	h.DB.Raw(`
SELECT COALESCE(SUM(ap.amount), 0)
FROM actual_payments ap
JOIN contracts c ON c.id = ap.contract_id
WHERE c.type <> 'sales' AND ap.pay_date >= ? AND ap.pay_date <= ?
`, yearStart, yearEnd).Scan(&yearExpense)

	c.JSON(http.StatusOK, gin.H{
		"asOf": today.Format("2006-01-02"),
		"contracts": gin.H{
			"total":      totals.TotalCount,
			"active":     totals.ActiveCount,
			"completed":  totals.CompletedCt,
			"terminated": totals.TerminatedCt,
		},
		"currentMonth": today.Format("2006-01"),
		"payable": gin.H{
			"dueUnpaidAmount": round2(monthExpenseDue),
			"paidAmount":      round2(monthExpensePaid),
			"unpaidSnapshot":  round2(monthExpenseDue),
		},
		"receivable": gin.H{
			"dueUncollectedAmount": round2(monthRecvDue),
			"collectedAmount":      round2(monthRecvCollected),
		},
		"byType": byType,
		"year": gin.H{
			"year":    y,
			"income":  round2(yearIncome),
			"expense": round2(yearExpense),
		},
	})
}

func round2(x float64) float64 {
	return float64(int64(x*100+0.5)) / 100
}
