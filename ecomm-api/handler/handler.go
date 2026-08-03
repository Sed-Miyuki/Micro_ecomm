package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/repo"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/server"
	"github.com/go-chi/chi/v5"
)

type handler struct{
	ctx		context.Context
	server 	*server.Server
}

func NewHandler(server *server.Server) *handler{
	return &handler{ctx: context.Background(),server: server}
}

func toRepoProduct(p ProductReq) *repo.Product{
	return &repo.Product{
		Name: p.Name,
		Image: p.Image,
		Category: p.Category,
		Description: p.Description,
		Rating: p.Rating,
		NumReviews: p.NumReviews,
		Price: p.Price,
		CountInStock: p.CountInStock,
	}
}
func toProductRes(product *repo.Product) ProductRes{
	return ProductRes{
		ID: product.ID,
		Name: product.Name,
		Image: product.Image,
		Category: product.Category,
		Description: product.Description,
		Rating: product.Rating,
		NumReviews: product.NumReviews,
		Price: product.Price,
		CountInStock: product.CountInStock,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}
}
func patchProductReq(product *repo.Product, p ProductReq) {
	if p.Name != "" {
		product.Name = p.Name
	}
	if p.Image != "" {
		product.Image = p.Image
	}
	if p.Category != "" {
		product.Category = p.Category
	}
	if p.Description != "" {
		product.Description = p.Description
	}
	if p.Rating != 0 {
		product.Rating = p.Rating
	}
	if p.NumReviews != 0 {
		product.NumReviews = p.NumReviews
	}
	if p.Price != 0 {
		product.Price = p.Price
	}
	if p.CountInStock != 0 {
		product.CountInStock = p.CountInStock
	}
	now := time.Now()
	product.UpdatedAt = &now
}
func toRepoOrder(o OrderReq) *repo.Order{
	return &repo.Order{
		PaymentMethod: o.PaymentMethod,
		TaxPrice: o.TaxPrice,
		TotalPrice: o.TotalPrice,
		ShippingPrice: o.ShippingPrice,
		Items: toRepoOrderItems(o.Items),
	}
}
func toRepoOrderItems(items []OrderItem) []repo.OrderItem{
	var res []repo.OrderItem
	for _,i:=range(items){
		res = append(res, repo.OrderItem{
			Name: i.Name,
			Quantity: i.Quantity,
			Image: i.Image,
			Price: i.Price,
			ProductId: i.ProductID,
		})
	}
	return res
}
func toOrderRes(o *repo.Order) OrderRes{
	return OrderRes{
		ID: o.ID,
		PaymentMethod: o.PaymentMethod,
		ShippingPrice: o.ShippingPrice,
		TaxPrice: o.TaxPrice,
		TotalPrice: o.TotalPrice,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Items: toOrderItems(o.Items),
	}
}
func toOrderItems(items []repo.OrderItem) []OrderItem{
	var res []OrderItem
	for _,i:=range(items){
		res = append(res, OrderItem{
			Name: i.Name,
			Quantity: i.Quantity,
			Image: i.Image,
			Price: i.Price,
			ProductID: i.ProductId,
		})
	}
	return res
}

func (h *handler) createProduct(w http.ResponseWriter,r *http.Request){
	var p ProductReq
	if err:=json.NewDecoder(r.Body).Decode(&p);err!=nil{
		http.Error(w,"error decoding request body",http.StatusBadRequest)
		return
	}

	product,err:=h.server.CreateProduct(h.ctx,toRepoProduct(p))
	if err!=nil{
		http.Error(w,"error creating product: %w",http.StatusInternalServerError)
		return 
	}
	res:=toProductRes(product)

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

//paramaterized query product/{id}
func (h *handler) getProduct(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	i,err:=strconv.ParseInt(id,10,64)
	if err!=nil{
		http.Error(w,"error parsing ID",http.StatusBadRequest)
		return
	}
	product,err:=h.server.GetProduct(h.ctx,i)
	if err!=nil{
		http.Error(w,"error getting product",http.StatusInternalServerError)
	}
	res:=toProductRes(product)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listProduct(w http.ResponseWriter,r *http.Request){
	product,err:=h.server.ListProducts(h.ctx)
	if err!=nil{
		http.Error(w,"error getting product",http.StatusInternalServerError)
	}
	var res []ProductRes
	for _,p:=range(product){
		res=append(res, toProductRes(&p))
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateProduct(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	i,err:=strconv.ParseInt(id,10,64)
	if err!=nil{
		http.Error(w,"error parsing ID",http.StatusBadRequest)
		return
	}
	var p ProductReq
	if err:=json.NewDecoder(r.Body).Decode(&p);err!=nil{
		http.Error(w,"error decoding request body",http.StatusBadRequest)
		return
	}
	product,err:=h.server.GetProduct(h.ctx,i)
	if err!=nil{
		http.Error(w,"error getting product",http.StatusInternalServerError)
		return
	}
	patchProductReq(product,p)
	updated,err:=h.server.UpdateProduct(h.ctx,product)
	if err!=nil{
		http.Error(w,"error updating product",http.StatusInternalServerError)
		return
	}
	res:=toProductRes(updated)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteProduct(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	i,err:=strconv.ParseInt(id,10,64)
	if err!=nil{
		http.Error(w,"error parsing ID",http.StatusBadRequest)
		return
	}
	if err:=h.server.DeleteProduct(h.ctx,i);err!=nil{
		http.Error(w, fmt.Sprintf("database error: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createOrder(w http.ResponseWriter,r *http.Request){
	var o OrderReq
	if err:=json.NewDecoder(r.Body).Decode(&o);err!=nil{
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	created,err:=h.server.CreateOrder(h.ctx,toRepoOrder(o))
	if err!=nil{
		http.Error(w,"internal server error",http.StatusInternalServerError)
		return
	}
	res:=toOrderRes(created)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) getOrder(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	i,err:=strconv.ParseInt(id,10,64)
	if err!=nil{
		panic(err)
	}
	order,err:=h.server.GetOrder(h.ctx,i)
	if err!=nil{
		http.Error(w,"internal server error",http.StatusInternalServerError)
		return
	}
	res:=toOrderRes(order)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listOrders(w http.ResponseWriter,r *http.Request){
	orders,err:=h.server.ListOrders(h.ctx)
	if err!=nil{
		http.Error(w,"internal server error",http.StatusInternalServerError)
		return
	}
	var res []OrderRes
	for _,i:=range(orders){
		res = append(res, toOrderRes(&i))
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteOrder(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	i,err:=strconv.ParseInt(id,10,64)
	if err!=nil{
		panic(err)
	}
	err=h.server.DeleteOrder(h.ctx,i)
	if err!=nil{
		http.Error(w,"internal server error",http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

