package entity

import (
	"time"

	"github.com/workdestiny/amlporn/config"
	"github.com/shopspring/decimal"
)

//GetViewRevenueModel view and guest model
type GetViewRevenueModel struct {
	View            int64           `json:"view"`
	ViewAmount      decimal.Decimal `json:"viewAmount"`
	GuestView       int64           `json:"guestView"`
	GuestViewAmount decimal.Decimal `json:"guestViewAmount"`
	All             int64           `json:"all"`
	AllAmount       decimal.Decimal `json:"allAmount"`
	Amount          string          `json:"amount"`
	Percent         float64         `json:"percent"`
}

//Sum GetViewModel
func (v GetViewRevenueModel) Sum() int64 {
	return v.View + v.GuestView
}

//AmountView GetViewModel
func (v GetViewRevenueModel) AmountView() decimal.Decimal {
	return decimal.New(int64(v.View), 0).Mul(config.RevenueRateView)
}

//AmountGuestView GetViewModel
func (v GetViewRevenueModel) AmountGuestView() decimal.Decimal {
	return decimal.New(int64(v.GuestView), 0).Mul(config.RevenueRateGuestView)
}

//AmountAll GetViewModel
func (v GetViewRevenueModel) AmountAll() decimal.Decimal {
	return v.AmountView().Add(v.AmountGuestView())
}

//AmountPercent GetViewModel
func (v GetViewRevenueModel) AmountPercent() float64 {
	am, _ := v.AmountAll().Float64()
	return (am * 100) / config.MinimumPay
}

// Bookbank User
type Bookbank struct {
	Name     string `json:"name"`
	Number   string `json:"number"`
	BankName string `json:"bankName"`
	Image    string `json:"image"`
}

// RevenueModel User
type RevenueModel struct {
	ID        string          `json:"name"`
	Gap       GapList         `json:"gap"`
	Time      string          `json:"time"`
	CreatedAt time.Time       `json:"createdAt"`
	Total     decimal.Decimal `json:"total"`
}

// RevenueStatus status
type RevenueStatus int

const (
	// Pending is status
	Pending RevenueStatus = iota
	// Approve is status
	Approve
	// Reject is status
	Reject
)

var mapRevenueStatusString = map[RevenueStatus]string{
	Pending: "Pending",
	Approve: "Approve",
	Reject:  "Reject",
}

func (x RevenueStatus) String() string {
	return mapRevenueStatusString[x]
}
