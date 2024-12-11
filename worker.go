package main

import (
	"context"
	"errors"
	"sync"
)

type UnprocessedOrderIterator interface {
	GetNextUnprocessedBooking(ctx context.Context) (*Order, error)
}

type unprocessedOrderIterator struct {
	orderStorage FirstUnprocessedOrderProvider
}

func NewUnprocessedOrderIterator(orderStorage FirstUnprocessedOrderProvider) UnprocessedOrderIterator {
	return &unprocessedOrderIterator{orderStorage}
}

func (uoi *unprocessedOrderIterator) GetNextUnprocessedBooking(ctx context.Context) (*Order, error) {
	for {
		order, err := uoi.orderStorage.GetFirstUnprocessedOrder(ctx)
		if errors.Is(err, OrderNotFoundError) {
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
	ProcessOrder(ctx context.Context, order *Order) error
}

type orderProcessor struct {
	roomBooker   RoomBooker
	orderUpdater OrderUpdater
}

func NewOrderProcessor(roomBooker RoomBooker, orderUpdater OrderUpdater) OrderProcessor {
	return &orderProcessor{roomBooker: roomBooker, orderUpdater: orderUpdater}
}

func (op *orderProcessor) ProcessOrder(ctx context.Context, order *Order) error {
	if err := op.roomBooker.Book(order); err != nil {
		return err
	}
	if _, err := op.orderUpdater.SetProcessed(ctx, order.ID, true); err != nil {
		return err
	}

	return nil
}
