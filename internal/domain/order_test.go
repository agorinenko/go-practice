package domain_test

import (
	"errors"
	"testing"

	"orders/internal/domain"
)

func validItems() []domain.OrderItem {
	return []domain.OrderItem{
		{ProductID: "p-1", Name: "Книга", Quantity: 2, Price: 1_000},
		{ProductID: "p-2", Name: "Ручка", Quantity: 5, Price: 50},
	}
}

func TestNewOrder(t *testing.T) {
	o, err := domain.NewOrder("o-1", "c-1", validItems())
	if err != nil {
		t.Fatalf("NewOrder: unexpected error: %v", err)
	}
	if o.ID != "o-1" {
		t.Errorf("ID = %q, want %q", o.ID, "o-1")
	}
	if o.CustomerID != "c-1" {
		t.Errorf("CustomerID = %q, want %q", o.CustomerID, "c-1")
	}
	if o.Status != domain.StatusNew {
		t.Errorf("Status = %q, want %q", o.Status, domain.StatusNew)
	}
	if want := int64(2*1_000 + 5*50); o.Total != want {
		t.Errorf("Total = %d, want %d", o.Total, want)
	}
	if o.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero value")
	}
}

func TestNewOrderErrors(t *testing.T) {
	cases := []struct {
		name       string
		customerID string
		items      []domain.OrderItem
		wantErr    error
	}{
		{"empty customer", "", validItems(), domain.ErrEmptyCustomer},
		{"no items", "c-1", nil, domain.ErrNoItems},
		{"zero quantity", "c-1", []domain.OrderItem{{ProductID: "p-1", Quantity: 0, Price: 10}}, domain.ErrInvalidItem},
		{"negative price", "c-1", []domain.OrderItem{{ProductID: "p-1", Quantity: 1, Price: -5}}, domain.ErrInvalidItem},
		{"empty product id", "c-1", []domain.OrderItem{{Quantity: 1, Price: 10}}, domain.ErrInvalidItem},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewOrder("o-1", tc.customerID, tc.items)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("NewOrder error = %v, want errors.Is %v", err, tc.wantErr)
			}
		})
	}
}

func TestRecalcTotal(t *testing.T) {
	o := &domain.Order{Items: validItems()}
	o.RecalcTotal()
	if want := int64(2*1_000 + 5*50); o.Total != want {
		t.Fatalf("RecalcTotal: Total = %d, want %d", o.Total, want)
	}
}

func TestApplyStatusValidTransitions(t *testing.T) {
	valid := []struct {
		from, to domain.OrderStatus
	}{
		{domain.StatusNew, domain.StatusValidated},
		{domain.StatusValidated, domain.StatusPaid},
		{domain.StatusNew, domain.StatusCancelled},
	}
	for _, tc := range valid {
		o := &domain.Order{Status: tc.from}
		if err := o.ApplyStatus(tc.to); err != nil {
			t.Errorf("ApplyStatus(%v → %v): unexpected error: %v", tc.from, tc.to, err)
		}
		if o.Status != tc.to {
			t.Errorf("ApplyStatus(%v → %v): Status = %v", tc.from, tc.to, o.Status)
		}
	}
}

func TestApplyStatusInvalidTransitions(t *testing.T) {
	invalid := []struct {
		from, to domain.OrderStatus
	}{
		{domain.StatusNew, domain.StatusPaid},       // пропуск validated
		{domain.StatusPaid, domain.StatusValidated}, // назад нельзя
		{domain.StatusValidated, domain.StatusNew},  // назад нельзя
		{domain.StatusCancelled, domain.StatusPaid}, // из cancelled никуда
		{domain.StatusPaid, domain.StatusCancelled}, // оплаченный не отменяется
	}
	for _, tc := range invalid {
		o := &domain.Order{Status: tc.from}
		if err := o.ApplyStatus(tc.to); !errors.Is(err, domain.ErrInvalidTransition) {
			t.Errorf("ApplyStatus(%v → %v) error = %v, want ErrInvalidTransition", tc.from, tc.to, err)
		}
		if o.Status != tc.from {
			t.Errorf("ApplyStatus(%v → %v): статус изменился на %v, должен остаться %v", tc.from, tc.to, o.Status, tc.from)
		}
	}
}

func TestValidate(t *testing.T) {
	valid := &domain.Order{ID: "o-1", CustomerID: "c-1", Items: validItems()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate(valid): unexpected error: %v", err)
	}

	noCustomer := &domain.Order{ID: "o-1", Items: validItems()}
	if err := noCustomer.Validate(); !errors.Is(err, domain.ErrEmptyCustomer) {
		t.Errorf("Validate(no customer) = %v, want ErrEmptyCustomer", err)
	}

	noItems := &domain.Order{ID: "o-1", CustomerID: "c-1"}
	if err := noItems.Validate(); !errors.Is(err, domain.ErrNoItems) {
		t.Errorf("Validate(no items) = %v, want ErrNoItems", err)
	}
}
