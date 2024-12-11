package worker

import (
	"applicationDesignTest/booking"
	"applicationDesignTest/orders"
	"applicationDesignTest/util"
	"context"
	"errors"
	"fmt"
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
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
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
}

type Worker interface {
	Work(ctx context.Context)
}

type worker struct {
	unprocessedOrderIterator UnprocessedOrderIterator
	orderProcessor           booking.OrderProcessor
	logger                   util.Logger
}

func NewWorker(orderProcessor booking.OrderProcessor, orderIterator UnprocessedOrderIterator, logger util.Logger) Worker {
	return &worker{orderProcessor: orderProcessor, unprocessedOrderIterator: orderIterator, logger: logger}
}

func (op *worker) Work(ctx context.Context) {
	defer op.logger.Log(util.Info, "Worker stopped")

	op.logger.Log(util.Info, "Worker started working")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			order, err := op.unprocessedOrderIterator.GetNextUnprocessedBooking(ctx)
			if err != nil {
				op.logger.Log(util.Error, fmt.Sprintf("Failed to get unprocessed order: %v", err))
				continue
			}

			err = op.orderProcessor.ProcessOrder(ctx, order)
			if errors.Is(err, booking.RoomUnavailableError) {
				op.logger.Log(util.Info, err.Error(), util.LogEnv{"order", *order})
				continue
			}
			if err != nil {
				op.logger.Log(util.Error, fmt.Sprintf("Failed to process order: %v", err))
				continue
			}
		}
	}
}
