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
	"applicationDesignTest/availability"
	"applicationDesignTest/book"
	"applicationDesignTest/orders"
	"applicationDesignTest/util"
	"applicationDesignTest/worker"
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

func main() {
	ctx := context.Background()
	logger := util.NewLogger(os.Stdout)

	orderStorage := orders.NewOrderStorage(make([]*orders.Order, InitialOrderCapacity))
	orderCreator := orders.NewOrderCreator(orderStorage)
	orderController := NewOrderController(orderCreator)
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", orderController.CreateOrderMethod)
	// todo method GET /orders/{id}

	logger.Log(util.Info, "Server listening on localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Log(util.Info, "Server closed")
	} else if err != nil {
		logger.Log(util.Error, fmt.Sprintf("Server failed: %s", err))
		os.Exit(1)
	}

	var availabilityData = []availability.RoomAvailability{
		{"reddison", "lux", date(2024, 1, 1), 1},
		{"reddison", "lux", date(2024, 1, 2), 1},
		{"reddison", "lux", date(2024, 1, 3), 1},
		{"reddison", "lux", date(2024, 1, 4), 1},
		{"reddison", "lux", date(2024, 1, 5), 0},
	}
	availabilityManager := availability.NewAvailabilityManagerInMemory(availabilityData, logger)
	roomBooker := book.NewRoomBooker(availabilityManager)
	unprocessedOrderIterator := worker.NewUnprocessedOrderIterator(orderStorage)
	orderProcessor := worker.NewOrderProcessor(roomBooker, orderStorage)
	worker := worker.NewWorker(orderProcessor, unprocessedOrderIterator)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go worker.Work(ctx, wg)
	wg.Wait()
}

type OrderController struct {
	orderCreator orders.OrderCreator
}

func NewOrderController(orderCreator orders.OrderCreator) *OrderController {
	return &OrderController{orderCreator}
}

func (oc *OrderController) CreateOrderMethod(w http.ResponseWriter, r *http.Request) {
	var newOrder orders.OrderDto
	json.NewDecoder(r.Body).Decode(&newOrder)

	orderDto, err := oc.orderCreator.CreateOrder(r.Context(), &newOrder)
	if err != nil {
		http.Error(w, "Hotel room is not available for selected dates", http.StatusInternalServerError)
		return
	}

	order := &Order{
		ID:        uint(orderDto.ID),
		HotelID:   string(orderDto.HotelID),
		RoomID:    string(orderDto.RoomID),
		UserEmail: orderDto.UserEmail,
		From:      orderDto.From,
		To:        orderDto.To,
		Processed: orderDto.Processed,
		Success:   orderDto.Success,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
