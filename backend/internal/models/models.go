package models

import (
	"time"

	"gorm.io/gorm"
)

type Contract struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	ContractNo   string         `json:"contractNo" gorm:"column:contract_no;uniqueIndex;size:64"`
	Title        string         `json:"title" gorm:"size:256"`
	SignedDate   time.Time      `json:"signedDate" gorm:"column:signed_date;type:date"`
	Type         string         `json:"type" gorm:"size:24"`
	Counterparty string         `json:"counterparty" gorm:"size:256"`
	TotalAmount  float64        `json:"totalAmount" gorm:"column:total_amount;type:numeric(18,2)"`
	PeriodStart  *time.Time     `json:"periodStart,omitempty" gorm:"column:period_start;type:date"`
	PeriodEnd    *time.Time     `json:"periodEnd,omitempty" gorm:"column:period_end;type:date"`
	Summary      string         `json:"summary" gorm:"type:text"`
	Status       string         `json:"status" gorm:"size:24"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	PaymentNodes []PaymentNode  `json:"-" gorm:"foreignKey:ContractID"`
}

func (Contract) TableName() string {
	return "contracts"
}

type PaymentNode struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	ContractID       uint           `json:"contractId" gorm:"column:contract_id;index"`
	NodeName         string         `json:"nodeName" gorm:"column:node_name;size:128"`
	TriggerCondition string         `json:"triggerCondition" gorm:"column:trigger_condition;type:text"`
	Amount           float64        `json:"amount" gorm:"type:numeric(18,2)"`
	PlannedDate      time.Time      `json:"plannedDate" gorm:"column:planned_date;type:date"`
	IsTriggered      bool           `json:"isTriggered" gorm:"column:is_triggered"`
	IsPaid           bool           `json:"isPaid" gorm:"column:is_paid"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

func (PaymentNode) TableName() string {
	return "payment_nodes"
}

type ActualPayment struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	ContractID uint           `json:"contractId" gorm:"column:contract_id;index"`
	NodeID     uint           `json:"nodeId" gorm:"column:node_id"`
	PayDate    time.Time      `json:"payDate" gorm:"column:pay_date;type:date"`
	Amount     float64        `json:"amount" gorm:"type:numeric(18,2)"`
	BankRef    string         `json:"bankRef" gorm:"column:bank_ref;size:128"`
	PayAccount string         `json:"payAccount" gorm:"column:pay_account;size:128"`
	CreatedAt  time.Time      `json:"createdAt"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ActualPayment) TableName() string {
	return "actual_payments"
}

type CollectionFollowup struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ContractID      uint           `json:"contractId" gorm:"column:contract_id;index"`
	NodeID          *uint          `json:"nodeId,omitempty" gorm:"column:node_id"`
	Follower        string         `json:"follower" gorm:"size:64"`
	FollowDate      time.Time      `json:"followDate" gorm:"column:follow_date;type:date"`
	Content         string         `json:"content" gorm:"type:text"`
	PromisedPayDate *time.Time     `json:"promisedPayDate,omitempty" gorm:"column:promised_pay_date;type:date"`
	CreatedAt       time.Time      `json:"createdAt"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (CollectionFollowup) TableName() string {
	return "collection_followups"
}
