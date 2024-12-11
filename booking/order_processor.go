package booking

import (
	"applicationDesignTest/orders"
	"context"
	"errors"
	"fmt"
)

type OrderProcessor interface {
	ProcessOrder(ctx context.Context, order *orders.Order) error
}

type orderProcessor struct {
	roomBooker   RoomBooker
	orderUpdater orders.OrderUpdater
}

func NewOrderProcessor(roomBooker RoomBooker, orderUpdater orders.OrderUpdater) OrderProcessor {
	return &orderProcessor{roomBooker: roomBooker, orderUpdater: orderUpdater}
}

func (op *orderProcessor) ProcessOrder(ctx context.Context, order *orders.Order) error {
	err := op.roomBooker.Book(order)
	if errors.Is(err, RoomUnavailableError) {
		if _, err := op.orderUpdater.SetProcessed(ctx, order.ID, false); err != nil {
			return err
		}
		return RoomUnavailableError
	}
	if err != nil {
		return fmt.Errorf("error when booking a room: %w", err)
	}

	if _, err := op.orderUpdater.SetProcessed(ctx, order.ID, true); err != nil {
		return err
	}

	return nil
}
