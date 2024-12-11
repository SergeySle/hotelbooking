package worker

import (
	"applicationDesignTest/book"
	"applicationDesignTest/orders"
	"applicationDesignTest/util"
	"context"
	"errors"
	"fmt"
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
	Work(ctx context.Context, wg *sync.WaitGroup) error
}

type worker struct {
	unprocessedOrderIterator UnprocessedOrderIterator
	orderProcessor           book.OrderProcessor
	logger                   util.Logger
}

func NewWorker(orderProcessor book.OrderProcessor, orderIterator UnprocessedOrderIterator, logger util.Logger) Worker {
	return &worker{orderProcessor: orderProcessor, unprocessedOrderIterator: orderIterator, logger: logger}
}

func (op *worker) Work(ctx context.Context, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
		op.logger.Log(util.Info, "Worker stopped")
	}()

	op.logger.Log(util.Info, "Worker started working")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			order, err := op.unprocessedOrderIterator.GetNextUnprocessedBooking(ctx)
			if err != nil {
				return err
			}

			err = op.orderProcessor.ProcessOrder(ctx, order)
			if errors.Is(err, book.RoomUnavailableError) {
				op.logger.Log(util.Info, err.Error(), util.LogEnv{"order", *order})
				continue
			}
			if err != nil {
				return fmt.Errorf("error processing order: %w", err)
			}
		}
	}
}
