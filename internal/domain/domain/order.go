package domain

import (
	"fmt"
	"time"
)

const (
	NewOrderStatus    = "new"
	PaidOrderStatus   = "paid"
	OtherOrderStatus  = "other"
	FailedOrderStatus = "failed"

	orderDescriptionTemplate = `CheckIT order:
Customer %s
Order item: %s
Price: %d$
Order time: %s`
)

type Order struct {
	ID           string
	PlanID       string
	UserID       string
	Status       string
	Amount       int
	Currency     string
	Description  string
	Transactions []*Transaction
	CreatedAt    time.Time
}

func (o *Order) Paid() bool {
	return o.Status == PaidOrderStatus
}

type Transaction struct {
	Status         string
	CreatedAt      time.Time
	AdditionalInfo string
}

func GenerateOrderDescription(name, orderItem string, amount int, orderTime time.Time) string {
	return fmt.Sprintf(
		orderDescriptionTemplate,
		name,
		orderItem,
		amount,
		orderTime.Format(time.RFC822),
	)
}
