package main

import (
	"fmt"
	"time"
)

type Booking struct {
	ID        uint      `json:"id"`
	HotelID   string    `json:"hotel_id"`
	RoomID    string    `json:"room_id"`
	UserEmail string    `json:"email"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
}

type RoomBooker interface {
	Book(order *Order) error
}

type roomBooker struct {
	availabilityManager AvailabilityManager
}

func NewRoomBooker(availabilityManager AvailabilityManager) RoomBooker {
	return &roomBooker{availabilityManager}
}

func (rr *roomBooker) Book(order *Order) error {
	daysToBook := daysBetween(order.From, order.To)

	err := rr.availabilityManager.UpdateAvailability(
		BookingRequest{BookingId(order.ID), HotelId(order.HotelID), RoomId(order.RoomID), daysToBook},
	)
	if err != nil {
		return fmt.Errorf("Can't book a room: %w", err)
	}

	return nil
}

func daysBetween(from time.Time, to time.Time) []time.Time {
	if from.After(to) {
		return nil
	}

	days := make([]time.Time, 0)
	for d := toDay(from); !d.After(toDay(to)); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}

	return days
}

func toDay(timestamp time.Time) time.Time {
	return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, time.UTC)
}
