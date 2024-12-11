package routing

import (
	"applicationDesignTest/controller"
	"github.com/go-chi/chi/v5"
)

type Routing struct {
	orderController *controller.OrderController
}

func NewRouting(orderController *controller.OrderController) *Routing {
	return &Routing{orderController: orderController}
}

func (ro Routing) Route(r *chi.Mux) {
	r.Post("/orders", ro.orderController.CreateOrder)
	r.Get("/orders/{id}", ro.orderController.GetOrder)
}
