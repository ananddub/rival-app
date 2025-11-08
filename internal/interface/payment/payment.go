package payment

import (
	"context"
	"time"
)

type Payment struct {
	ID             int64
	UserID         int64
	OrderID        string
	PaymentID      *string
	Amount         float64
	Currency       string
	Status         string
	PaymentMethod  *string
	PaymentGateway *string
	PaymentType    string
	ReferenceID    *int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Repository interface {
	CreatePayment(ctx context.Context, payment *Payment) (int64, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdatePaymentStatus(ctx context.Context, orderID, paymentID, status string) error
	GetUserPayments(ctx context.Context, userID int64) ([]*Payment, error)
}

type Service interface {
	InitiatePayment(ctx context.Context, userID int64, amount float64, paymentType string, referenceID int64) (string, error)
	VerifyPayment(ctx context.Context, orderID, paymentID string) error
	GetPaymentHistory(ctx context.Context, userID int64) ([]*Payment, error)
}
