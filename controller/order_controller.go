package controller

import (
	"applicationDesignTest/orders"
	"applicationDesignTest/web"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

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

	order := web.NewOrder(orderDto)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (oc *OrderController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var newOrder web.OrderRequest
	json.NewDecoder(r.Body).Decode(&newOrder)

	orderDto, err := oc.orderCreator.CreateOrder(r.Context(), newOrder.ToOrderData())
	if err != nil {
		http.Error(w, "Can't create order", http.StatusInternalServerError)
		return
	}

	order := web.NewOrder(orderDto)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
