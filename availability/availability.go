package availability

import (
	"applicationDesignTest/util"
	"errors"
	"fmt"
	"sync"
	"time"
)

type HotelId string
type RoomId string
type BookingId uint

type RoomAvailability struct {
	HotelID string
	RoomID  string
	Date    time.Time
	Quota   int
}

type BookingRequest struct {
	BookingId BookingId
	HotelId   HotelId
	RoomId    RoomId
	DateRange []time.Time
}

func (b BookingRequest) Validate() error {
	if b.BookingId == 0 {
		return errors.New("BookingId is required")
	}

	if b.HotelId == "" {
		return errors.New("HotelId is required")
	}

	if b.RoomId == "" {
		return errors.New("RoomId is required")
	}

	if len(b.DateRange) == 0 {
		return errors.New("DateRange is required")
	}

	for _, d := range b.DateRange {
		if d.Hour() != 0 || d.Minute() != 0 || d.Second() != 0 {
			return errors.New("Hour, minute and second must be zero")
		}
	}

	for i := 1; i < len(b.DateRange); i++ {
		if b.DateRange[i].Sub(b.DateRange[i-1]) != time.Hour*24 {
			return fmt.Errorf("Invalid date range, error at %d and %d", i-1, i)
		}
	}

	return nil
}

type BookingSlot struct {
	HotelId HotelId
	RoomId  RoomId
	Date    time.Time
}

type AvailabilityData struct {
	Quota      int
	BookingIds *util.Set[BookingId]
}

func NewAvailabilityData(quota int) AvailabilityData {
	return AvailabilityData{Quota: quota, BookingIds: util.NewSet[BookingId](quota)}
}

type AvailabilityManager interface {
	UpdateAvailability(request BookingRequest) error
}

type availabilityManagerInMemory struct {
	availabilityStorage map[BookingSlot]AvailabilityData
	logger              util.Logger
	mu                  sync.RWMutex
}

func NewAvailabilityManagerInMemory(data []RoomAvailability, logger util.Logger) AvailabilityManager {
	m := make(map[BookingSlot]AvailabilityData, len(data))

	for _, roomAvailability := range data {
		m[BookingSlot{HotelId(roomAvailability.HotelID), RoomId(roomAvailability.RoomID), roomAvailability.Date}] = NewAvailabilityData(roomAvailability.Quota)
	}

	return &availabilityManagerInMemory{m, logger, sync.RWMutex{}}
}

func (ami *availabilityManagerInMemory) isAvailable(request BookingRequest) error {
	daysToBook := request.DateRange

	unavailableDays := make(map[time.Time]struct{})
	for _, day := range daysToBook {
		unavailableDays[day] = struct{}{}
	}

	ami.mu.RLock()
	for _, dayToBook := range daysToBook {
		slot := BookingSlot{request.HotelId, request.RoomId, dayToBook}

		availability, found := ami.availabilityStorage[slot]
		if !found || availability.Quota-availability.BookingIds.Size() < 1 {
			continue
		}

		delete(unavailableDays, dayToBook)
	}
	ami.mu.RUnlock()

	if len(unavailableDays) != 0 {
		ami.logger.Log(util.Info,
			fmt.Sprintf("Hotel room is not available for selected dates"),
			util.LogEnv{"request", request},
			util.LogEnv{"unavailable days", unavailableDays},
		)

		return errors.New("Hotel room is not available for selected dates")
	}

	return nil
}

func (ami *availabilityManagerInMemory) UpdateAvailability(request BookingRequest) error {
	if err := ami.isAvailable(request); err != nil {
		return fmt.Errorf("Can't book a room: %w", err)
	}

	ami.mu.Lock()
	for _, dayToBook := range request.DateRange {
		slot := BookingSlot{request.HotelId, request.RoomId, dayToBook}
		availabilityData, found := ami.availabilityStorage[slot]
		if !found {
			return errors.New("No availability data found")
		}
		availabilityData.BookingIds.Add(request.BookingId)

		ami.availabilityStorage[slot] = availabilityData
	}
	ami.mu.Unlock()

	return nil
}
