package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

var r *chi.Mux

func RegisterRoutes(handler *handler) *chi.Mux{
	r=chi.NewRouter()
	r.Route("/products",func(r chi.Router) {
		r.Post("/",handler.createProduct)
		r.Get("/",handler.listProduct)
		r.Route("/{id}",func(r chi.Router) {
			r.Get("/",handler.getProduct)
			r.Patch("/",handler.updateProduct)
			r.Delete("/",handler.deleteProduct)
		})
	})
	r.Route("/orders",func(r chi.Router) {
		r.Post("/",handler.createOrder)
		r.Get("/",handler.listOrders)
		r.Route("/{id}",func(r chi.Router) {
			r.Get("/",handler.getOrder)
			r.Delete("/",handler.deleteOrder)
		})
	})
	return r
}

//*chi.mux implements http.Handler interface
func Start(addr string) error{
	return http.ListenAndServe(addr,r)
}