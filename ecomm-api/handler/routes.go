package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

var r *chi.Mux

func RegisterRoutes(handler *handler) *chi.Mux {
	r = chi.NewRouter()
	tokenMaker := handler.TokenMaker
	r.Route("/products", func(r chi.Router) {
		r.With(GetAdminMiddlewareFunc(tokenMaker)).Post("/", handler.createProduct)
		r.Get("/", handler.listProduct)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.getProduct)
			r.Group(func(r chi.Router) {
				r.Use(GetAdminMiddlewareFunc(tokenMaker))
				r.Patch("/", handler.updateProduct)
				r.Delete("/", handler.deleteProduct)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(GetAuthMiddlewareFunc(tokenMaker))
		r.Get("/myorder", handler.getOrder)
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", handler.createOrder)
			r.With(GetAdminMiddlewareFunc(tokenMaker)).Get("/", handler.listOrders)
			r.Route("/{id}", func(r chi.Router) {
				r.Delete("/", handler.deleteOrder)
			})
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handler.createUser)
		r.Post("/login", handler.loginUser)
		r.Group(func(r chi.Router) {
			r.Use(GetAdminMiddlewareFunc(tokenMaker))
			r.Get("/", handler.listUsers)
			r.Route("/{id}", func(r chi.Router) {
				r.Delete("/", handler.deleteUser)
			})
		})
		r.Group(func(r chi.Router) {
			r.Use(GetAuthMiddlewareFunc(tokenMaker))
			r.Patch("/", handler.updateUser)
			r.Post("/logout", handler.LogoutUser)
		})
	})
	
	r.Post("/tokens/renew", handler.RenewAccessToken)

	r.Group(func(r chi.Router) {
		r.Use(GetAuthMiddlewareFunc(tokenMaker))
		r.Route("/tokens", func(r chi.Router) {
			r.Post("/revoke", handler.RevokeSession)
		})
	})
	return r
}

// *chi.mux implements http.Handler interface
func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
