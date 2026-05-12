package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"contractpayment/internal/models"

	"gorm.io/gorm"
)

func verifyContract(tx *gorm.DB, cid uint) (*models.Contract, bool) {
	var row models.Contract
	if err := tx.First(&row, cid).Error; err != nil {
		return nil, false
	}
	return &row, true
}

func (h *Routes) listNodes(c *gin.Context) {
	cid, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, ex := verifyContract(h.DB, cid); !ex {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	var list []models.PaymentNode
	if err := h.DB.Where("contract_id = ?", cid).Order("planned_date, id").Find(&list).Error; err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	var cont models.Contract
	if err := h.DB.First(&cont, cid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "contract load failed"})
		return
	}
	today := ShanghaiToday()
	out := make([]gin.H, 0, len(list))
	for _, n := range list {
		level := AlertLevel(cont.Type, n.PlannedDate, n.IsPaid, today)
		out = append(out, gin.H{
			"id":               n.ID,
			"contractId":       n.ContractID,
			"nodeName":         n.NodeName,
			"triggerCondition": n.TriggerCondition,
			"amount":           n.Amount,
			"plannedDate":      n.PlannedDate.Format("2006-01-02"),
			"isTriggered":      n.IsTriggered,
			"isPaid":           n.IsPaid,
			"alertLevel":       level,
		})
	}
	c.JSON(http.StatusOK, out)
}

type nodeBody struct {
	NodeName         string  `json:"nodeName"`
	TriggerCondition string  `json:"triggerCondition"`
	Amount           float64 `json:"amount"`
	PlannedDate      string  `json:"plannedDate"`
	IsTriggered      bool   `json:"isTriggered"`
}

func (h *Routes) createNode(c *gin.Context) {
	cid, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, ex := verifyContract(h.DB, cid); !ex {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	var body nodeBody
	if err := c.ShouldBindJSON(&body); err != nil || body.NodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	pd, err := parseDateOnly(body.PlannedDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad plannedDate"})
		return
	}
	row := models.PaymentNode{
		ContractID:       cid,
		NodeName:         body.NodeName,
		TriggerCondition: body.TriggerCondition,
		Amount:           body.Amount,
		PlannedDate:      pd,
		IsTriggered:      body.IsTriggered,
	}
	if err := h.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, row)
}

func (h *Routes) updateNode(c *gin.Context) {
	cid, ok := parseID(c, "id")
	if !ok {
		return
	}
	nid, ok := parseID(c, "nodeId")
	if !ok {
		return
	}
	var row models.PaymentNode
	if err := h.DB.First(&row, nid).Error; err != nil || row.ContractID != cid {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	var body nodeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.NodeName != "" {
		row.NodeName = body.NodeName
	}
	row.TriggerCondition = body.TriggerCondition
	row.Amount = body.Amount
	if body.PlannedDate != "" {
		pd, err := parseDateOnly(body.PlannedDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad plannedDate"})
			return
		}
		row.PlannedDate = pd
	}
	row.IsTriggered = body.IsTriggered
	if err := h.DB.Save(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *Routes) deleteNode(c *gin.Context) {
	cid, ok := parseID(c, "id")
	if !ok {
		return
	}
	nid, ok := parseID(c, "nodeId")
	if !ok {
		return
	}
	res := h.DB.Where("contract_id = ? AND id = ?", cid, nid).Delete(&models.PaymentNode{})
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

type followBody struct {
	NodeID          *uint   `json:"nodeId,omitempty"`
	Follower        string  `json:"follower"`
	FollowDate      string  `json:"followDate"`
	Content         string  `json:"content"`
	PromisedPayDate *string `json:"promisedPayDate,omitempty"`
}

func (h *Routes) listFollowups(c *gin.Context) {
	cid, ok := parseID(c, "id")
	if !ok {
		return
	}
	if _, ex := verifyContract(h.DB, cid); !ex {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	var list []models.CollectionFollowup
	if err := h.DB.Where("contract_id = ?", cid).Order("follow_date DESC, id DESC").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *Routes) createFollowup(c *gin.Context) {
	cid, ok := parseID(c, "id")
	if !ok {
		return
	}
	cont, ex := verifyContract(h.DB, cid)
	if !ex {
		c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		return
	}
	var body followBody
	if err := c.ShouldBindJSON(&body); err != nil || body.Follower == "" || body.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if cont.Type != "sales" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "followups apply to sales contracts only"})
		return
	}
	fd, err := parseDateOnly(body.FollowDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad followDate"})
		return
	}
	row := models.CollectionFollowup{
		ContractID: cid,
		NodeID:     body.NodeID,
		Follower:   body.Follower,
		FollowDate: fd,
		Content:    body.Content,
	}
	if body.PromisedPayDate != nil && *body.PromisedPayDate != "" {
		pp, err := parseDateOnly(*body.PromisedPayDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad promisedPayDate"})
			return
		}
		row.PromisedPayDate = &pp
	}
	if err := h.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, row)
}

type payBody struct {
	ContractID uint    `json:"contractId"`
	NodeID     uint    `json:"nodeId"`
	PayDate    string  `json:"payDate"`
	Amount     float64 `json:"amount"`
	BankRef    string  `json:"bankRef"`
	PayAccount string  `json:"payAccount"`
}

func (h *Routes) createPayment(c *gin.Context) {
	var body payBody
	if err := c.ShouldBindJSON(&body); err != nil || body.ContractID == 0 || body.NodeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	pd, err := parseDateOnly(body.PayDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad payDate"})
		return
	}

	var node models.PaymentNode
	if err := h.DB.First(&node, body.NodeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if node.ContractID != body.ContractID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node does not belong to contract"})
		return
	}
	if node.IsPaid {
		c.JSON(http.StatusConflict, gin.H{"error": "node already paid"})
		return
	}
	amt := body.Amount
	if amt <= 0 {
		amt = node.Amount
	}
	pay := models.ActualPayment{
		ContractID: body.ContractID,
		NodeID:     body.NodeID,
		PayDate:    pd,
		Amount:     amt,
		BankRef:    body.BankRef,
		PayAccount: body.PayAccount,
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&pay).Error; err != nil {
			return err
		}
		node.IsPaid = true
		node.IsTriggered = true
		return tx.Save(&node).Error
	})
	if err != nil {
		log.Println("createPayment", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "persist failed"})
		return
	}
	c.JSON(http.StatusCreated, pay)
}

func (h *Routes) listPayments(c *gin.Context) {
	q := h.DB.Model(&models.ActualPayment{}).Order("pay_date DESC, id DESC").Limit(1000)
	if cidStr := c.Query("contractId"); cidStr != "" {
		q = q.Where("contract_id = ?", cidStr)
	}
	var rows []models.ActualPayment
	if err := q.Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
