package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/repo"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/server"
	"github.com/Sed-Miyuki/Micro_ecomm/token"
	"github.com/Sed-Miyuki/Micro_ecomm/util"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type handler struct{
	ctx				context.Context
	server 			*server.Server
	TokenMaker		*token.JWTMaker
}

func NewHandler(server *server.Server,secretkey string) *handler{
	return &handler{ctx: context.Background(),server: server,TokenMaker: token.NewJWTMaker(secretkey),}
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
func patchUserReq(user *repo.User,u UserReq){
	if u.Name != "" {
		user.Name = u.Name
	}
	if u.Email != "" {
		user.Email = u.Email
	}
	if u.Password != "" {
		hashed, err := util.HashPassword(u.Password)
		if err != nil {
			panic(err)
		}
		user.Password = hashed
	}
	if u.IsAdmin {
		user.IsAdmin = u.IsAdmin
	}
	now := time.Now()
	user.UpdatedAt = &now
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
func toRepoUser(u UserReq) *repo.User{
	return &repo.User{
		Name: u.Name,
		Email: u.Email,
		Password: u.Password,
		IsAdmin: u.IsAdmin,
	}
}
func toUserRes(u *repo.User) UserRes{
	return UserRes{
		Name: u.Name,
		Email: u.Email,
		IsAdmin: u.IsAdmin,
	} 
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
	if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            http.Error(w, "product not found", http.StatusNotFound)
            return
        }
        http.Error(w, "error getting product", http.StatusInternalServerError)
        return
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
		return
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
	claims:=r.Context().Value(authkey{}).(*token.UserClaims)
	so:=toRepoOrder(o)
	so.UserID=claims.ID
	created,err:=h.server.CreateOrder(h.ctx,so)
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
	claims:=r.Context().Value(authkey{}).(*token.UserClaims)
	order,err:=h.server.GetOrder(h.ctx,claims.ID)
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

func (h *handler) createUser(w http.ResponseWriter,r *http.Request){
	var u UserReq
	if err:=json.NewDecoder(r.Body).Decode(&u);err!=nil{
		http.Error(w,"bad request",http.StatusInternalServerError)
		return
	}

	//hash password
	hahsed,err:=util.HashPassword(u.Password)
	if err!=nil{
		http.Error(w,"error hashing password",http.StatusInternalServerError)
		return
	}
	u.Password=hahsed

	created,err:=h.server.CreateUser(h.ctx,toRepoUser(u))
	if err!=nil{
		http.Error(w,"error creating user",http.StatusInternalServerError)
		return
	}
	res:=toUserRes(created)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listUsers(w http.ResponseWriter,_ *http.Request){
	users,err:=h.server.ListUsers(h.ctx)
	if err!=nil{
		http.Error(w,"error listing users",http.StatusInternalServerError)
		return
	}
	var res ListUserRes
	for _,i:=range(users){
		res.Users = append(res.Users, toUserRes(&i))
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateUser(w http.ResponseWriter,r *http.Request){
	var u UserReq
	if err:=json.NewDecoder(r.Body).Decode(&u);err!=nil{
		http.Error(w,"error decoding request body",http.StatusBadRequest)
		return
	}
	claims:=r.Context().Value(authkey{}).(*token.UserClaims)
	user,err:=h.server.GetUser(h.ctx,claims.Email)
	if err!=nil{
		http.Error(w,"error getting user",http.StatusInternalServerError)
		return
	}
	patchUserReq(user,u)
	if u.Email==""{
		u.Email=claims.Email
	}
	updated,err:=h.server.UpdateUser(h.ctx,user)
	if err!=nil{
		http.Error(w,"error updating user",http.StatusInternalServerError)
		return
	}
	res:=toUserRes(updated)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteUser(w http.ResponseWriter,r *http.Request){
	id:=chi.URLParam(r,"id")
	i,err:=strconv.ParseInt(id,10,64)
	if err!=nil{
		http.Error(w,"error parsing id",http.StatusBadRequest)
		return
	}
	err=h.server.DeleteUser(h.ctx,i)
	if err!=nil{
		http.Error(w,"error deleting user",http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) loginUser(w http.ResponseWriter,r *http.Request){
	var u LoginUserReq
	if err:=json.NewDecoder(r.Body).Decode(&u);err!=nil{
		http.Error(w,"error decoding request body",http.StatusBadRequest)
		return
	}

	gu,err:=h.server.GetUser(h.ctx,u.Email)
	if err!=nil{
		http.Error(w,"error getting user",http.StatusInternalServerError)
		return
	}

	err=util.CheckPassword(u.Password,gu.Password)
	if err!=nil{
		http.Error(w,"wrong username or password",http.StatusUnauthorized)
		return
	}

	accesstoken,accessclaims,err:=h.TokenMaker.CreateToken(gu.ID,gu.Email,gu.IsAdmin,15*time.Minute)
	if err!=nil{
		http.Error(w,"error creating access token",http.StatusInternalServerError)
		return
	}
	refreshtoken,refreshclaims,err:=h.TokenMaker.CreateToken(gu.ID,gu.Email,gu.IsAdmin,24*time.Hour)
	if err!=nil{
		http.Error(w,"error creating refresh token",http.StatusInternalServerError)
		return
	}

	session,err:=h.server.CreateSession(h.ctx,&repo.Session{
		ID: refreshclaims.RegisteredClaims.ID,
		UserEmail: gu.Email,
		RefreshToken: refreshtoken,
		IsRevoked: false,
		ExpiresAt: refreshclaims.ExpiresAt.Time,
	})
	
	if err!=nil{
		http.Error(w,"error creating session",http.StatusInternalServerError)
		return
	}

	res:=LoginUserRes{
		SessionID: session.ID,
		AccessToken: accesstoken,
		RefreshToken: refreshtoken,
		AccessTokenExpiresAt: accessclaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshclaims.ExpiresAt.Time,
		User: toUserRes(gu),
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) LogoutUser(w http.ResponseWriter,r *http.Request){
	claims,ok:=r.Context().Value(authkey{}).(*token.UserClaims)
	if !ok || claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
	err:=h.server.DeleteSession(h.ctx,claims.RegisteredClaims.ID)
	if err!=nil{
		http.Error(w,"error deleting session",http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) RenewAccessToken(w http.ResponseWriter,r *http.Request){
	var req RenewAccessTokenReq
	if err:=json.NewDecoder(r.Body).Decode(&req);err!=nil{
		http.Error(w,"error decoding request body",http.StatusBadRequest)
		return
	}
	refreshClaims,err:=h.TokenMaker.VerifyToken(req.RefreshToken)
	if err!=nil{
		http.Error(w,"error verifying token",http.StatusUnauthorized)
		return
	}
	session,err:=h.server.GetSession(h.ctx,refreshClaims.RegisteredClaims.ID)
	if err!=nil{
		http.Error(w,"missing session id",http.StatusInternalServerError)
		return
	}
	if session.IsRevoked{
		http.Error(w,"session revoked",http.StatusUnauthorized)
		return
	}
	if session.UserEmail!=refreshClaims.Email{
		http.Error(w,"invalid session",http.StatusUnauthorized)
		return
	}
	accessToken,accessClaims,err:=h.TokenMaker.CreateToken(refreshClaims.ID,refreshClaims.Email,refreshClaims.IsAdmin,15*time.Minute)
	if err!=nil{
		http.Error(w,"error creating token",http.StatusInternalServerError)
		return
	}
	res:=RenewAccessTokenRes{
		AccessToken: accessToken,
		AccessTokenExpiresAt: accessClaims.ExpiresAt.Time,
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) RevokeSession(w http.ResponseWriter,r *http.Request){
	claims,ok:=r.Context().Value(authkey{}).(*token.UserClaims)
	if !ok || claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
	err:=h.server.RevokeSession(h.ctx,claims.RegisteredClaims.ID)
	if err!=nil{
		http.Error(w,"error revoking session",http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}