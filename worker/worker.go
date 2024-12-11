package worker

import (
	"applicationDesignTest/book"
	"applicationDesignTest/orders"
	"context"
	"errors"
	"sync"
	"time"
)

type UnprocessedOrderIterator interface {
	GetNextUnprocessedBooking(ctx context.Context) (*orders.Order, error)
}

type unprocessedOrderIterator struct {
	orderStorage orders.FirstUnprocessedOrderProvider
}

func NewUnprocessedOrderIterator(orderStorage orders.FirstUnprocessedOrderProvider) UnprocessedOrderIterator {
	return &unprocessedOrderIterator{orderStorage}
}

func (uoi *unprocessedOrderIterator) GetNextUnprocessedBooking(ctx context.Context) (*orders.Order, error) {
	for {
		order, err := uoi.orderStorage.GetFirstUnprocessedOrder(ctx)
		if errors.Is(err, orders.OrderNotFoundError) {
			time.Sleep(time.Second)
			continue
		}
		if err != nil {
			return nil, err
		}

		return order, nil
	}
}

type Worker interface {
	ProcessOrders(ctx context.Context, wg *sync.WaitGroup) error
}

type worker struct {
	unprocessedOrderIterator UnprocessedOrderIterator
	orderProcessor           OrderProcessor
}

func NewWorker(orderProcessor OrderProcessor, orderIterator UnprocessedOrderIterator) Worker {
	return &worker{orderProcessor: orderProcessor, unprocessedOrderIterator: orderIterator}
}

func (op *worker) ProcessOrders(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			order, err := op.unprocessedOrderIterator.GetNextUnprocessedBooking(ctx)
			if err != nil {
				return err
			}

			if err := op.orderProcessor.ProcessOrder(ctx, order); err != nil {
				return err
			}
		}
	}
}

type OrderProcessor interface {
	ProcessOrder(ctx context.Context, order *orders.Order) error
}

type orderProcessor struct {
	roomBooker   book.RoomBooker
	orderUpdater orders.OrderUpdater
}

func NewOrderProcessor(roomBooker book.RoomBooker, orderUpdater orders.OrderUpdater) OrderProcessor {
	return &orderProcessor{roomBooker: roomBooker, orderUpdater: orderUpdater}
}

func (op *orderProcessor) ProcessOrder(ctx context.Context, order *orders.Order) error {
	if err := op.roomBooker.Book(order); err != nil {
		return err
	}
	if _, err := op.orderUpdater.SetProcessed(ctx, order.ID, true); err != nil {
		return err
	}

	return nil
}
