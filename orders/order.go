package orders

import (
	"applicationDesignTest/availability"
	"context"
	"errors"
	"fmt"
	"time"
)

type OrderData struct {
	HotelID   availability.HotelId
	RoomID    availability.RoomId
	UserEmail string
	From      time.Time
	To        time.Time
}

type Order struct {
	ID OrderId
	*OrderData
	Processed bool
	Success   bool
}

type OrderId uint

type OrderPersister interface {
	Persist(ctx context.Context, order *Order) (*Order, error)
}

type OrderUpdater interface {
	SetProcessed(ctx context.Context, orderId OrderId, success bool) (*Order, error)
}

type FirstUnprocessedOrderProvider interface {
	GetFirstUnprocessedOrder(ctx context.Context) (*Order, error)
}

type OrderCreator interface {
	CreateOrder(ctx context.Context, order *OrderData) (*Order, error)
}

type OrderStorage interface {
	OrderPersister
	OrderUpdater
	FirstUnprocessedOrderProvider
	OrderProvider
}

type OrderProvider interface {
	GetById(ctx context.Context, orderId OrderId) (*Order, error)
}

var OrderNotFoundError = errors.New("order not found")

type orderCreator struct {
	orderPersister OrderPersister
}

func NewOrderCreator(orderPersister OrderPersister) OrderCreator {
	return &orderCreator{orderPersister}
}

func (oc *orderCreator) CreateOrder(ctx context.Context, orderData *OrderData) (*Order, error) {
	order := &Order{OrderData: orderData, Processed: false, Success: false}
	order, err := oc.orderPersister.Persist(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("could not persist order: %v", err)
	}

	return order, nil
}

type orderStorageInMemory struct {
	orders []*Order
	maxId  uint
}

func NewOrderStorageInMemory(orders []*Order) *orderStorageInMemory {
	return &orderStorageInMemory{orders: orders}
}

func (s *orderStorageInMemory) Persist(ctx context.Context, order *Order) (*Order, error) {
	s.maxId++
	order.ID = OrderId(s.maxId)
	s.orders = append(s.orders, order)

	return order, nil
}

func (s *orderStorageInMemory) SetProcessed(ctx context.Context, orderId OrderId, success bool) (*Order, error) {
	for _, order := range s.orders {
		if order.ID == orderId {
			order.Processed = true
			order.Success = success

			return order, nil
		}
	}

	return nil, fmt.Errorf("order not found")
}

func (s *orderStorageInMemory) GetFirstUnprocessedOrder(ctx context.Context) (*Order, error) {
	for _, order := range s.orders {
		if !order.Processed {
			return order, nil
		}
	}

	return nil, OrderNotFoundError
}

func (s *orderStorageInMemory) GetById(ctx context.Context, orderId OrderId) (*Order, error) {
	for _, order := range s.orders {
		if order.ID == orderId {
			return order, nil
		}
	}

	return nil, OrderNotFoundError
}
