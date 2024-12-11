package web

import (
	"applicationDesignTest/availability"
	"applicationDesignTest/orders"
	"time"
)

type Order struct {
	ID        uint      `json:"id"`
	HotelID   string    `json:"hotel_id"`
	RoomID    string    `json:"room_id"`
	UserEmail string    `json:"email"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	Processed bool      `json:"processed"`
	Success   bool      `json:"success"`
}

func NewOrder(order *orders.Order) *Order {
	return &Order{
		ID:        uint(order.ID),
		HotelID:   string(order.HotelID),
		RoomID:    string(order.RoomID),
		UserEmail: order.UserEmail,
		From:      order.From,
		To:        order.To,
		Processed: order.Processed,
		Success:   order.Success,
	}
}

type OrderRequest struct {
	HotelID   availability.HotelId `json:"hotel_id"`
	RoomID    availability.RoomId  `json:"room_id"`
	UserEmail string               `json:"email"`
	From      time.Time            `json:"from"`
	To        time.Time            `json:"to"`
}

func (o OrderRequest) ToOrderData() *orders.OrderData {
	return &orders.OrderData{
		HotelID:   o.HotelID,
		RoomID:    o.RoomID,
		UserEmail: o.UserEmail,
		From:      o.From,
		To:        o.To,
	}
}
