package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"contractpayment/internal/models"

	"gorm.io/gorm"
)

type Routes struct {
	DB *gorm.DB
}

func Register(r gin.IRouter, db *gorm.DB) {
	h := &Routes{DB: db}
	api := r.Group("/api")

	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	api.GET("/health", h.health)

	api.GET("/contracts", h.listContracts)
	api.POST("/contracts", h.createContract)
	api.GET("/contracts/:id", h.getContract)
	api.PUT("/contracts/:id", h.updateContract)
	api.DELETE("/contracts/:id", h.deleteContract)

	api.GET("/contracts/:id/nodes", h.listNodes)
	api.POST("/contracts/:id/nodes", h.createNode)
	api.PUT("/contracts/:id/nodes/:nodeId", h.updateNode)
	api.DELETE("/contracts/:id/nodes/:nodeId", h.deleteNode)

	api.GET("/contracts/:id/followups", h.listFollowups)
	api.POST("/contracts/:id/followups", h.createFollowup)

	api.GET("/payments", h.listPayments)
	api.POST("/payments", h.createPayment)

	api.GET("/dashboard/upcoming", h.dashboardUpcoming)
	api.GET("/reports/summary", h.reportSummary)
}

func (h *Routes) health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil {
		c.String(http.StatusServiceUnavailable, "db closed")
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.String(http.StatusServiceUnavailable, "db ping failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "time": time.Now().In(shanghaiTZ).Format(time.RFC3339)})
}

func parseID(c *gin.Context, key string) (uint, bool) {
	s := c.Param(key)
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(n), true
}

func (h *Routes) listContracts(c *gin.Context) {
	var list []models.Contract
	q := h.DB.Model(&models.Contract{})
	if kw := strings.TrimSpace(c.Query("q")); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("title ILIKE ? OR contract_no ILIKE ? OR counterparty ILIKE ?", like, like, like)
	}
	if t := strings.TrimSpace(c.Query("type")); t != "" {
		q = q.Where("type = ?", t)
	}
	if st := strings.TrimSpace(c.Query("status")); st != "" {
		q = q.Where("status = ?", st)
	}
	if err := q.Order("signed_date DESC, id DESC").Find(&list).Error; err != nil {
		log.Println("listContracts", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, list)
}

type contractBody struct {
	ContractNo   string  `json:"contractNo"`
	Title        string  `json:"title"`
	SignedDate   string  `json:"signedDate"`
	Type         string  `json:"type"`
	Counterparty string  `json:"counterparty"`
	TotalAmount  float64 `json:"totalAmount"`
	PeriodStart  *string `json:"periodStart,omitempty"`
	PeriodEnd    *string `json:"periodEnd,omitempty"`
	Summary      string  `json:"summary"`
	Status       string  `json:"status"`
}

func parseDateOnly(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(s), shanghaiTZ)
}

func (h *Routes) createContract(c *gin.Context) {
	var body contractBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if body.ContractNo == "" || body.Title == "" || body.Type == "" || body.Counterparty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}
	sd, err := parseDateOnly(body.SignedDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad signedDate"})
		return
	}
	row := models.Contract{
		ContractNo:   body.ContractNo,
		Title:        body.Title,
		SignedDate:   sd,
		Type:         body.Type,
		Counterparty: body.Counterparty,
		TotalAmount:  body.TotalAmount,
		Summary:      body.Summary,
		Status:       body.Status,
	}
	if row.Status == "" {
		row.Status = "active"
	}
	if body.PeriodStart != nil && *body.PeriodStart != "" {
		t, e := parseDateOnly(*body.PeriodStart)
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad periodStart"})
			return
		}
		row.PeriodStart = &t
	}
	if body.PeriodEnd != nil && *body.PeriodEnd != "" {
		t, e := parseDateOnly(*body.PeriodEnd)
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad periodEnd"})
			return
		}
		row.PeriodEnd = &t
	}
	if err := h.DB.Create(&row).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "contract number exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, row)
}

func (h *Routes) getContract(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var row models.Contract
	if err := h.DB.First(&row, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *Routes) updateContract(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var body contractBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var row models.Contract
	if err := h.DB.First(&row, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	if body.Title != "" {
		row.Title = body.Title
	}
	if body.ContractNo != "" && body.ContractNo != row.ContractNo {
		row.ContractNo = body.ContractNo
	}
	if body.SignedDate != "" {
		sd, err := parseDateOnly(body.SignedDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad signedDate"})
			return
		}
		row.SignedDate = sd
	}
	if body.Type != "" {
		row.Type = body.Type
	}
	if body.Counterparty != "" {
		row.Counterparty = body.Counterparty
	}
	row.TotalAmount = body.TotalAmount
	row.Summary = body.Summary
	if body.Status != "" {
		row.Status = body.Status
	}
	row.PeriodStart = nil
	row.PeriodEnd = nil
	if body.PeriodStart != nil && *body.PeriodStart != "" {
		t, e := parseDateOnly(*body.PeriodStart)
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad periodStart"})
			return
		}
		row.PeriodStart = &t
	}
	if body.PeriodEnd != nil && *body.PeriodEnd != "" {
		t, e := parseDateOnly(*body.PeriodEnd)
		if e != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad periodEnd"})
			return
		}
		row.PeriodEnd = &t
	}
	if err := h.DB.Save(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *Routes) deleteContract(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	res := h.DB.Delete(&models.Contract{}, id)
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	c.Status(http.StatusNoContent)
}
