// Package domain содержит доменную модель заказа.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// OrderStatus — жизненный цикл заказа.
type OrderStatus string

// Статусы жизненного цикла заказа.
const (
	StatusNew       OrderStatus = "new"
	StatusValidated OrderStatus = "validated"
	StatusPaid      OrderStatus = "paid"
	StatusCancelled OrderStatus = "cancelled"
)

// OrderItem — позиция заказа. Деньги — int64 копеек, никогда не float64.
type OrderItem struct {
	ProductID string
	Name      string
	Quantity  int
	Price     int64
}

// Order — заказ.
type Order struct {
	ID         string
	CustomerID string
	Items      []OrderItem
	Status     OrderStatus
	Total      int64 // в копейках
	CreatedAt  time.Time
}

// Sentinel-ошибки домена: вызывающий код проверяет их через errors.Is.
var (
	// ErrEmptyCustomer — пустой идентификатор покупателя.
	ErrEmptyCustomer = errors.New("orders: customer id is empty")
	// ErrNoItems — заказ без позиций.
	ErrNoItems = errors.New("orders: order has no items")
	// ErrInvalidItem — некорректная позиция заказа.
	ErrInvalidItem = errors.New("orders: invalid item")
	// ErrOrderNotFound — заказ не найден.
	ErrOrderNotFound = errors.New("orders: order not found")
	// ErrInvalidTransition — недопустимый переход статуса.
	ErrInvalidTransition = errors.New("orders: invalid status transition")
)

// validTransitions — допустимые переходы статусов заказа.
var validTransitions = map[OrderStatus]map[OrderStatus]struct{}{
	StatusNew:       {StatusValidated: {}, StatusCancelled: {}},
	StatusValidated: {StatusPaid: {}},
}

// NewOrder создаёт заказ в статусе StatusNew.
func NewOrder(id, customerID string, items []OrderItem) (*Order, error) {
	var order = &Order{
		ID:         id,
		CustomerID: customerID,
		Items:      items,
		Status:     StatusNew,
		CreatedAt:  time.Now(),
	}
	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("validate order %s: %w", order.ID, err)
	}
	order.RecalcTotal()
	return order, nil
}

// Validate проверяет инварианты заказа: непустой CustomerID, есть позиции,
// каждая позиция валидна (ProductID непустой, Quantity > 0, Price >= 0).
func (o *Order) Validate() error {
	if strings.TrimSpace(o.CustomerID) == "" {
		return fmt.Errorf("%w: CustomerID is empty", ErrEmptyCustomer)
	}

	if len(o.Items) == 0 {
		return fmt.Errorf("%w: order %q: items is empty", ErrNoItems, o.CustomerID)
	}

	for _, item := range o.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return fmt.Errorf("%w: product %q: id is empty", ErrInvalidItem, item.ProductID)
		}

		if item.Quantity <= 0 {
			return fmt.Errorf("%w: product %q: quantity must be > 0", ErrInvalidItem, item.ProductID)
		}

		if item.Price < 0 {

			return fmt.Errorf("%w: product %q: price must be >= 0", ErrInvalidItem, item.ProductID)
		}
	}
	return nil
}

// RecalcTotal пересчитывает Total заказа по позициям.
func (o *Order) RecalcTotal() {
	var total int64
	for _, item := range o.Items {
		total += item.Amount()
	}

	o.Total = total
}

// ApplyStatus применяет новый статус заказа.
// Допустимые переходы: new → validated → paid; new → cancelled.
// Недопустимый переход — ErrInvalidTransition с указанием обоих статусов.
func (o *Order) ApplyStatus(next OrderStatus) error {
	if _, ok := validTransitions[o.Status][next]; ok {
		o.Status = next
		return nil
	}
	return fmt.Errorf("%w: %q → %q", ErrInvalidTransition, o.Status, next)
}

// Amount — сумма позиции (для удобства вызывающего кода).
func (i OrderItem) Amount() int64 {
	return int64(i.Quantity) * i.Price
}
