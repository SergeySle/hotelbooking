package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type OrderDto struct {
	HotelID   string    `json:"hotel_id"`
	RoomID    string    `json:"room_id"`
	UserEmail string    `json:"email"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
}

type Order struct {
	ID OrderId `json:"id"`
	*OrderDto
	Processed bool `json:"processed"`
	Success   bool `json:"success"`
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
	CreateOrder(ctx context.Context, order *OrderDto) (*Order, error)
}

type OrderStorage interface {
	OrderPersister
	OrderUpdater
	FirstUnprocessedOrderProvider
}

var OrderNotFoundError = errors.New("order not found")

type orderCreator struct {
	orderPersister OrderPersister
}

func NewOrderCreator(orderPersister OrderPersister) OrderCreator {
	return &orderCreator{orderPersister}
}

func (oc *orderCreator) CreateOrder(ctx context.Context, orderDto *OrderDto) (*Order, error) {
	order := &Order{OrderDto: orderDto, Processed: false, Success: false}
	order, err := oc.orderPersister.Persist(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("could not persist order: %v", err)
	}

	return order, nil
}

type orderStorage struct {
	orders []*Order
	maxId  uint
}

func NewOrderStorage(orders []*Order) *orderStorage {
	return &orderStorage{orders: orders}
}

func (s *orderStorage) Persist(ctx context.Context, order *Order) (*Order, error) {
	s.maxId++
	order.ID = OrderId(s.maxId)
	s.orders = append(s.orders, order)

	return order, nil
}

func (s *orderStorage) SetProcessed(ctx context.Context, orderId OrderId, success bool) (*Order, error) {
	for _, order := range s.orders {
		if order.ID == orderId {
			order.Processed = true
			order.Success = success

			return order, nil
		}
	}

	return nil, fmt.Errorf("order not found")
}

func (s *orderStorage) GetFirstUnprocessedOrder(ctx context.Context) (*Order, error) {
	for _, order := range s.orders {
		if !order.Processed {
			return order, nil
		}
	}

	return nil, OrderNotFoundError
}
