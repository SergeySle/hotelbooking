// Ниже реализован сервис бронирования номеров в отеле. В предметной области
// выделены два понятия: Order — заказ, который включает в себя даты бронирования
// и контакты пользователя, и RoomAvailability — количество свободных номеров на
// конкретный день.
//
// Задание:
// - провести рефакторинг кода с выделением слоев и абстракций
// - применить best-practices там где это имеет смысл
// - исправить имеющиеся в реализации логические и технические ошибки и неточности
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const InitialOrderCapacity = 1000

type Order2 struct {
	ID        uint      `json:"id"`
	HotelID   string    `json:"hotel_id"`
	RoomID    string    `json:"room_id"`
	UserEmail string    `json:"email"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
}

func main() {
	ctx := context.Background()
	logger := NewLogger()

	orderStorage := NewOrderStorage(make([]*Order, InitialOrderCapacity))
	orderCreator := NewOrderCreator(orderStorage)
	orderController := NewOrderController(orderCreator)
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", orderController.CreateOrderMethod)

	logger.Log(Info, "Server listening on localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Log(Info, "Server closed")
	} else if err != nil {
		logger.Log(Error, fmt.Sprintf("Server failed: %s", err))
		os.Exit(1)
	}

	var availability = []RoomAvailability{
		{"reddison", "lux", date(2024, 1, 1), 1},
		{"reddison", "lux", date(2024, 1, 2), 1},
		{"reddison", "lux", date(2024, 1, 3), 1},
		{"reddison", "lux", date(2024, 1, 4), 1},
		{"reddison", "lux", date(2024, 1, 5), 0},
	}
	availabilityManager := NewAvailabilityManagerInMemory(availability, logger)
	roomBooker := NewRoomBooker(availabilityManager)
	unprocessedOrderIterator := NewUnprocessedOrderIterator(orderStorage)
	orderProcessor := NewOrderProcessor(roomBooker, orderStorage)
	worker := NewWorker(orderProcessor, unprocessedOrderIterator)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go worker.ProcessOrders(ctx, wg)
	wg.Wait()
}

type OrderController struct {
	orderCreator OrderCreator
}

func NewOrderController(orderCreator OrderCreator) *OrderController {
	return &OrderController{orderCreator}
}

func (oc *OrderController) CreateOrderMethod(w http.ResponseWriter, r *http.Request) {
	var newOrder OrderDto
	json.NewDecoder(r.Body).Decode(&newOrder)

	order, err := oc.orderCreator.CreateOrder(r.Context(), &newOrder)
	if err != nil {
		http.Error(w, "Hotel room is not available for selected dates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
