package server

import (
	"context"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/repo"
)

type Server struct{
	pgs *repo.PgsRepo
}

func NewServer(pgs *repo.PgsRepo) *Server{
	return &Server{pgs:pgs}
}

func (s *Server) CreateProduct(ctx context.Context,p *repo.Product) (*repo.Product,error){
	return s.pgs.CreateProduct(ctx,p)
}

func (s *Server) GetProduct(ctx context.Context,id int64) (*repo.Product,error){
	return s.pgs.GetProduct(ctx,id)
}

func (s *Server) ListProducts(ctx context.Context) ([]repo.Product,error){
	return s.pgs.ListProducts(ctx)
}

func (s *Server) UpdateProduct(ctx context.Context,p *repo.Product) (*repo.Product,error){
	return s.pgs.UpdateProduct(ctx,p)
}

func (s *Server) DeleteProduct(ctx context.Context,id int64) error{
	return s.pgs.DeleteProduct(ctx,id)
}

func (s *Server) CreateOrder(ctx context.Context,o *repo.Order) (*repo.Order,error){
	return s.pgs.CreateOrder(ctx,o)
}

func (s *Server) GetOrder(ctx context.Context,id int64) (*repo.Order,error){
	return s.pgs.GetOrder(ctx,id)
}

func (s *Server) ListOrders(ctx context.Context) ([]repo.Order,error){
	return s.pgs.ListOrders(ctx)
}

func (s *Server) DeleteOrder(ctx context.Context,id int64) error{
	return s.pgs.DeleteOrder(ctx,id)
}

func (s *Server) CreateUser(ctx context.Context,u *repo.User) (*repo.User,error){
	return s.pgs.CreateUser(ctx,u)
}

func (s *Server) GetUser(ctx context.Context,email string) (*repo.User,error){
	return s.pgs.GetUser(ctx,email)
}

func (s *Server) ListUsers(ctx context.Context) ([]repo.User,error){
	return s.pgs.ListUsers(ctx)
}

func (s *Server) UpdateUser(ctx context.Context,u *repo.User) (*repo.User,error){
	return s.pgs.UpdateUser(ctx,u)
}

func (s *Server) DeleteUser(ctx context.Context,id int64) error{
	return s.pgs.DeleteUser(ctx,id)
}

func (s *Server) CreateSession(ctx context.Context,se *repo.Session) (*repo.Session,error){
	return s.pgs.CreateSession(ctx,se)
}

func (s *Server) GetSession(ctx context.Context,id string) (*repo.Session,error){
	return s.pgs.GetSession(ctx,id)
}

func (s *Server) RevokeSession(ctx context.Context,id string) error{
	return s.pgs.RevokeSession(ctx,id)
}

func (s *Server) DeleteSession(ctx context.Context,id string) error{
	return s.pgs.DeleteSession(ctx,id)
}