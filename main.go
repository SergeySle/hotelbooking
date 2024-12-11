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
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"strconv"
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

func main() {
	ctx := context.Background()
	logger := util.NewLogger(log.Default())

	orderStorage := orders.NewOrderStorage(make([]*orders.Order, InitialOrderCapacity))
	orderCreator := orders.NewOrderCreator(orderStorage)
	orderController := NewOrderController(orderCreator, orderStorage)

	r := chi.NewRouter()
	r.Post("/orders", orderController.CreateOrderMethod)
	r.Get("/orders/{id}", orderController.GetOrder)

	logger.Log(util.Info, "Server listening on localhost:8080")
	err := http.ListenAndServe(":8080", r)
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
	orderCreator  orders.OrderCreator
	orderProvider orders.OrderProvider
}

func NewOrderController(orderCreator orders.OrderCreator, orderProvider orders.OrderProvider) *OrderController {
	return &OrderController{orderCreator, orderProvider}
}

func (oc *OrderController) GetOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderDto, err := oc.orderProvider.GetById(r.Context(), orders.OrderId(id))
	if err != nil {
		http.Error(w, fmt.Sprintf("can't get order by id: %w", err), http.StatusNotFound)
		return
	}

	order := NewOrder(orderDto)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (oc *OrderController) CreateOrderMethod(w http.ResponseWriter, r *http.Request) {
	var newOrder orders.OrderData
	json.NewDecoder(r.Body).Decode(&newOrder)

	orderDto, err := oc.orderCreator.CreateOrder(r.Context(), &newOrder)
	if err != nil {
		http.Error(w, "Hotel room is not available for selected dates", http.StatusInternalServerError)
		return
	}

	order := NewOrder(orderDto)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
