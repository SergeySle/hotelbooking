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
	"applicationDesignTest/controller"
	"applicationDesignTest/orders"
	"applicationDesignTest/routing"
	"applicationDesignTest/util"
	"applicationDesignTest/worker"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const InitialOrderCapacity = 1000

var availabilityData = []availability.RoomAvailability{
	{"reddison", "lux", date(2024, 1, 1), 1},
	{"reddison", "lux", date(2024, 1, 2), 1},
	{"reddison", "lux", date(2024, 1, 3), 1},
	{"reddison", "lux", date(2024, 1, 4), 1},
	{"reddison", "lux", date(2024, 1, 5), 0},
}

func main() {
	ctx := context.Background()
	logger := util.NewLogger(log.Default())

	orderStorage := orders.NewOrderStorage(make([]*orders.Order, 0, InitialOrderCapacity))
	orderCreator := orders.NewOrderCreator(orderStorage)
	orderController := controller.NewOrderController(orderCreator, orderStorage)

	router := routing.NewRouting(orderController)
	r := chi.NewRouter()
	router.Route(r)

	availabilityManager := availability.NewAvailabilityManagerInMemory(availabilityData, logger)
	roomBooker := book.NewRoomBooker(availabilityManager, logger)
	unprocessedOrderIterator := worker.NewUnprocessedOrderIterator(orderStorage)
	orderProcessor := worker.NewOrderProcessor(roomBooker, orderStorage)
	worker := worker.NewWorker(orderProcessor, unprocessedOrderIterator, logger)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go worker.Work(ctx, wg)
	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Log(util.Info, "Server listening on localhost:8080")
		err := http.ListenAndServe(":8080", r)
		if errors.Is(err, http.ErrServerClosed) {
			logger.Log(util.Info, "Server closed")
		} else if err != nil {
			logger.Log(util.Error, fmt.Sprintf("Server failed: %s", err))
			os.Exit(1)
		}
	}()
	wg.Wait()
}

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
