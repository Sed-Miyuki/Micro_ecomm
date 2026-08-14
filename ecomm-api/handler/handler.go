package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/pb"
	"github.com/Sed-Miyuki/Micro_ecomm/token"
	"github.com/Sed-Miyuki/Micro_ecomm/util"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type handler struct {
	ctx        context.Context
	client     pb.EcommClient
	TokenMaker *token.JWTMaker
}

func NewHandler(client pb.EcommClient, secretkey string) *handler {
	return &handler{
		ctx:        context.Background(),
		client:     client,
		TokenMaker: token.NewJWTMaker(secretkey),
	}
}

func (h *handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var p ProductReq
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	product, err := h.client.CreateProduct(h.ctx, toPBProductReq(p))
	if err != nil {
		http.Error(w, "error creating product", http.StatusInternalServerError)
		return
	}
	res := toProductRes(product)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// paramaterized query product/{id}
func (h *handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}
	product, err := h.client.GetProduct(h.ctx, &pb.ProductReq{Id: i})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "error getting product", http.StatusInternalServerError)
		return
	}
	res := toProductRes(product)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listProduct(w http.ResponseWriter, r *http.Request) {
	lpr, err := h.client.ListProducts(h.ctx, &pb.ProductReq{})
	if err != nil {
		http.Error(w, "error getting product", http.StatusInternalServerError)
		return
	}
	var res []ProductRes
	for _, p := range lpr.GetProducts() {
		res = append(res, toProductRes(p))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}
	var p ProductReq
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	p.ID = i
	updated, err := h.client.UpdateProduct(h.ctx, toPBProductReq(p))
	if err != nil {
		http.Error(w, "error updating product", http.StatusInternalServerError)
		return
	}
	res := toProductRes(updated)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}
	if _, err := h.client.DeleteProduct(h.ctx, &pb.ProductReq{Id: i}); err != nil {
		http.Error(w, fmt.Sprintf("database error: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var o OrderReq
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	claims := r.Context().Value(authkey{}).(*token.UserClaims)
	po := toPBOrderReq(o)
	po.UserId = claims.ID
	po.UserEmail = claims.Email
	created, err := h.client.CreateOrder(h.ctx, po)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	res := toOrderRes(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) getOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authkey{}).(*token.UserClaims)
	order, err := h.client.GetOrder(h.ctx, &pb.OrderReq{UserId: claims.ID})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	res := toOrderRes(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.client.ListOrders(h.ctx, &pb.OrderReq{})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	var res []OrderRes
	for _, o := range orders.Orders {
		res = append(res, toOrderRes(o))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateOrderStatus(w http.ResponseWriter,r *http.Request){
	claims:=r.Context().Value(authkey{}).(*token.UserClaims)
	var o OrderReq
	if err:=json.NewDecoder(r.Body).Decode(&o);err!=nil{
		http.Error(w,"error decoding request body",http.StatusBadRequest)
		return
	}
	status,err:=toPBOrderStatus(OrderStatus(o.Status))
	if err!=nil{
		http.Error(w,"invalid status",http.StatusBadRequest)
		return
	}
	res,err:=h.client.UpdateOrderStatus(h.ctx,&pb.OrderReq{
		Id: o.ID,
		UserId: claims.ID,
		UserEmail: claims.Email,
		Status: status,
	})
	if err!=nil{
		http.Error(w,"failed to update order status",http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}
	_, err = h.client.DeleteOrder(h.ctx, &pb.OrderReq{Id: i})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	var u UserReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}

	//hash password
	hashed, err := util.HashPassword(u.Password)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	u.Password = hashed

	created, err := h.client.CreateUser(h.ctx, toPBUserReq(u))
	if err != nil {
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}
	res := toUserRes(created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) listUsers(w http.ResponseWriter, _ *http.Request) {
	users, err := h.client.ListUsers(h.ctx, &pb.UserReq{})
	if err != nil {
		http.Error(w, "error listing users", http.StatusInternalServerError)
		return
	}
	var res ListUserRes
	for _, i := range users.Users {
		res.Users = append(res.Users, toUserRes(i))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var u UserReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	claims := r.Context().Value(authkey{}).(*token.UserClaims)
	u.Email = claims.Email
	updated, err := h.client.UpdateUser(h.ctx, toPBUserReq(u))
	if err != nil {
		http.Error(w, "error updating user", http.StatusInternalServerError)
		return
	}
	res := toUserRes(updated)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing id", http.StatusBadRequest)
		return
	}
	_, err = h.client.DeleteUser(h.ctx, &pb.UserReq{Id: i})
	if err != nil {
		http.Error(w, "error deleting user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var u LoginUserReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	ur, err := h.client.GetUser(h.ctx, &pb.UserReq{Email: u.Email})
	if err != nil {
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}

	err = util.CheckPassword(u.Password, ur.GetPassword())
	if err != nil {
		http.Error(w, "wrong username or password", http.StatusUnauthorized)
		return
	}

	accesstoken, accessclaims, err := h.TokenMaker.CreateToken(ur.GetId(), ur.GetEmail(), ur.GetIsAdmin(), 15*time.Minute)
	if err != nil {
		http.Error(w, "error creating access token", http.StatusInternalServerError)
		return
	}
	refreshtoken, refreshclaims, err := h.TokenMaker.CreateToken(ur.GetId(), ur.Email, ur.GetIsAdmin(), 24*time.Hour)
	if err != nil {
		http.Error(w, "error creating refresh token", http.StatusInternalServerError)
		return
	}

	session, err := h.client.CreateSession(h.ctx, &pb.SessionReq{
		Id:           refreshclaims.RegisteredClaims.ID,
		UserEmail:    ur.GetEmail(),
		RefreshToken: refreshtoken,
		IsRevoked:    false,
		ExpiresAt:    timestamppb.New(refreshclaims.ExpiresAt.Time),
	})

	if err != nil {
		http.Error(w, "error creating session", http.StatusInternalServerError)
		return
	}

	res := LoginUserRes{
		SessionID:             session.GetId(),
		AccessToken:           accesstoken,
		RefreshToken:          refreshtoken,
		AccessTokenExpiresAt:  accessclaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshclaims.ExpiresAt.Time,
		User:                  toUserRes(ur),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(authkey{}).(*token.UserClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_, err := h.client.DeleteSession(h.ctx, &pb.SessionReq{Id: claims.RegisteredClaims.ID})
	if err != nil {
		http.Error(w, "error deleting session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	var req RenewAccessTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	refreshClaims, err := h.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "error verifying token", http.StatusUnauthorized)
		return
	}
	session, err := h.client.GetSession(h.ctx, &pb.SessionReq{Id: refreshClaims.RegisteredClaims.ID})
	if err != nil {
		http.Error(w, "missing session id", http.StatusInternalServerError)
		return
	}
	if session.IsRevoked {
		http.Error(w, "session revoked", http.StatusUnauthorized)
		return
	}
	if session.UserEmail != refreshClaims.Email {
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, 15*time.Minute)
	if err != nil {
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}
	res := RenewAccessTokenRes{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessClaims.ExpiresAt.Time,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(authkey{}).(*token.UserClaims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_, err := h.client.RevokeSession(h.ctx, &pb.SessionReq{Id: claims.RegisteredClaims.ID})
	if err != nil {
		http.Error(w, "error revoking session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}