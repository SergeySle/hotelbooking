package book

import (
	"applicationDesignTest/availability"
	"applicationDesignTest/orders"
	"applicationDesignTest/util"
	"errors"
	"fmt"
	"time"
)

type Booking struct {
	ID        uint
	HotelID   string
	RoomID    string
	UserEmail string
	From      time.Time
	To        time.Time
}

var RoomUnavailableError = errors.New("room is unavailable")

type RoomBooker interface {
	Book(order *orders.Order) error
}

type roomBooker struct {
	logger              util.Logger
	availabilityManager availability.AvailabilityManager
}

func NewRoomBooker(availabilityManager availability.AvailabilityManager, logger util.Logger) RoomBooker {
	return &roomBooker{logger, availabilityManager}
}

func (rr *roomBooker) Book(order *orders.Order) error {
	daysToBook := daysBetween(order.From, order.To)

	err := rr.availabilityManager.UpdateAvailability(
		availability.BookingRequest{availability.BookingId(order.ID), availability.HotelId(order.HotelID), availability.RoomId(order.RoomID), daysToBook},
	)
	if err != nil {
		rr.logger.Log(util.Info, fmt.Sprintf("Can't book a room: %w", err))
		return RoomUnavailableError
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
