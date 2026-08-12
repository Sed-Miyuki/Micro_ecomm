package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/pb"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/repo"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mapRepoError translates repo-layer errors into gRPC status errors so that
// the REST API can map them to proper HTTP status codes (e.g. 404 Not Found).
func mapRepoError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.NotFound, "resource not found")
	}
	return status.Error(codes.Internal, err.Error())
}

type Server struct {
	repo *repo.PgsRepo
	pb.UnimplementedEcommServer
}

func NewServer(repo *repo.PgsRepo) *Server {
	return &Server{repo: repo}
}

func (s *Server) CreateProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	pr, err := s.repo.CreateProduct(ctx, torepoProduct(p))
	if err != nil {
		return nil, err
	}
	return toPBProductRes(pr), nil
}

func (s *Server) GetProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	pr, err := s.repo.GetProduct(ctx, p.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return toPBProductRes(pr), nil
}

func (s *Server) ListProducts(ctx context.Context, p *pb.ProductReq) (*pb.ListProductRes, error) {
	lps, err := s.repo.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	lpr := make([]*pb.ProductRes, 0, len(lps))
	for _, lp := range lps {
		lpr = append(lpr, toPBProductRes(lp))
	}

	return &pb.ListProductRes{
		Products: lpr,
	}, nil
}

func (s *Server) UpdateProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	product, err := s.repo.GetProduct(ctx, p.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	patchProductReq(product, p)
	pr, err := s.repo.UpdateProduct(ctx, product)
	if err != nil {
		return nil, mapRepoError(err)
	}

	return toPBProductRes(pr), nil
}

func (s *Server) DeleteProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	err := s.repo.DeleteProduct(ctx, p.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.ProductRes{}, nil
}

func (s *Server) CreateOrder(ctx context.Context, o *pb.OrderReq) (*pb.OrderRes, error) {
	order, err := s.repo.CreateOrder(ctx, torepoOrder(o))
	if err != nil {
		return nil, err
	}
	order.Status=repo.Pending

	_,err=s.repo.EnqueueNotificationEvent(ctx,&repo.NotificationEvent{
		UserEmail: o.GetUserEmail(),
		OrderStatus: order.Status,
		OrderID: order.ID,
		Attempts: 0,
	})
	if err!=nil{
		return nil,err
	}

	return toPBOrderRes(order), nil
}

func (s *Server) GetOrder(ctx context.Context, o *pb.OrderReq) (*pb.OrderRes, error) {
	order, err := s.repo.GetOrder(ctx, o.GetUserId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return toPBOrderRes(order), nil
}

func (s *Server) ListOrders(ctx context.Context, o *pb.OrderReq) (*pb.ListOrderRes, error) {
	orders, err := s.repo.ListOrders(ctx)
	if err != nil {
		return nil, err
	}

	lor := make([]*pb.OrderRes, 0, len(orders))
	for _, order := range orders {
		lor = append(lor, toPBOrderRes(order))
	}

	return &pb.ListOrderRes{
		Orders: lor,
	}, nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context,o *pb.OrderReq) (*pb.OrderRes,error){
	order,err:=s.repo.GetOrderStatusByID(ctx,o.GetId())
	if err!=nil{
		return nil,err
	}
	
	if o.GetUserId()!=order.UserID{
		return nil,fmt.Errorf("Order %d doesn't belong to the user %d",o.GetId(),o.GetUserId())
	}

	rOrderStatus:=repo.OrderStatus(strings.ToLower(o.GetStatus().String()))
	if rOrderStatus==order.Status{
		return nil,fmt.Errorf("Order status is already %s",order.Status)
	}
	order.Status=rOrderStatus
	or,err:=s.repo.UpdateOrderStatus(ctx,order)
	if err!=nil{
		return nil,err
	}

	_,err=s.repo.EnqueueNotificationEvent(ctx,&repo.NotificationEvent{
		UserEmail: o.GetUserEmail(),
		OrderStatus: order.Status,
		OrderID: order.ID,
		Attempts: 0,
	})
	if err!=nil{
		return nil,err
	}

	return toPBOrderRes(or),nil
}

func (s *Server) DeleteOrder(ctx context.Context, o *pb.OrderReq) (*pb.OrderRes, error) {
	err := s.repo.DeleteOrder(ctx, o.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.OrderRes{}, nil
}

func (s *Server) CreateUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	user, err := s.repo.CreateUser(ctx, torepoUser(u))
	if err != nil {
		return nil, err
	}

	return toPBUserRes(user), nil
}

func (s *Server) GetUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	user, err := s.repo.GetUser(ctx, u.GetEmail())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return toPBUserRes(user), nil
}

func (s *Server) ListUsers(ctx context.Context, u *pb.UserReq) (*pb.ListUserRes, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	lur := make([]*pb.UserRes, 0, len(users))
	for _, user := range users {
		lur = append(lur, toPBUserRes(user))
	}

	return &pb.ListUserRes{
		Users: lur,
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	user, err := s.repo.GetUser(ctx, u.GetEmail())
	if err != nil {
		return nil, mapRepoError(err)
	}

	patchUserReq(user, u)
	ur, err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, mapRepoError(err)
	}

	return toPBUserRes(ur), nil
}

func (s *Server) DeleteUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	err := s.repo.DeleteUser(ctx, u.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.UserRes{}, nil
}

func (s *Server) CreateSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	sess, err := s.repo.CreateSession(ctx, &repo.Session{
		ID:           sr.GetId(),
		UserEmail:    sr.GetUserEmail(),
		RefreshToken: sr.GetRefreshToken(),
		IsRevoked:    sr.GetIsRevoked(),
		ExpiresAt:    sr.GetExpiresAt().AsTime(),
	})
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.SessionRes{
		Id:           sess.ID,
		UserEmail:    sess.UserEmail,
		RefreshToken: sess.RefreshToken,
		IsRevoked:    sess.IsRevoked,
		ExpiresAt:    timestamppb.New(sess.ExpiresAt),
	}, nil
}

func (s *Server) GetSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	sess, err := s.repo.GetSession(ctx, sr.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.SessionRes{
		Id:           sess.ID,
		UserEmail:    sess.UserEmail,
		RefreshToken: sess.RefreshToken,
		IsRevoked:    sess.IsRevoked,
		ExpiresAt:    timestamppb.New(sess.ExpiresAt),
	}, nil
}

func (s *Server) RevokeSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	err := s.repo.RevokeSession(ctx, sr.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.SessionRes{}, nil
}

func (s *Server) DeleteSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	err := s.repo.DeleteSession(ctx, sr.GetId())
	if err != nil {
		return nil, mapRepoError(err)
	}

	return &pb.SessionRes{}, nil
}

func (s *Server) ListNotificationEvents(ctx context.Context, lnr *pb.ListNotificationEventsReq) (*pb.ListNotificationEventsRes,error){
	notificationEvents,err:=s.repo.ListNotificationEvents(ctx)
	if err!=nil{
		return nil,err
	}

	lners:=make([]*pb.NotificationEvent,0,len(notificationEvents))
	for _,ne:=range notificationEvents{
		lners=append(lners, &pb.NotificationEvent{
			Id: ne.ID,
			UserEmail: ne.UserEmail,
			OrderStatus: toPBOrderStatus(ne.OrderStatus),
			OrderId: ne.OrderID,
			StateId: ne.StateID,
			Attempts: ne.Attempts,
		})
	}
	return &pb.ListNotificationEventsRes{
		Events: lners,
	},nil
}

func (s *Server) UpdateNotificationEvent(ctx context.Context,unr *pb.UpdateNotificationEventReq) (*pb.UpdateNotificationEventRes,error){
	var resposeType repo.NotificationResponseType
	switch unr.ResponseType{
	case pb.NotificationResponseType_SUCCESS:
		resposeType=repo.NotificationSucess
	case pb.NotificationResponseType_FAILURE:
		resposeType=repo.NotificationFailure
	default:
		return nil,fmt.Errorf("invalid response type: %v",unr.ResponseType)
	}

	succeeded,err:=s.repo.UpdateNotificationEvent(ctx,
		&repo.NotificationEvent{
			ID: unr.GetId(),
			StateID: unr.GetStateId(),
		},
		&repo.NotificationState{
			Message: unr.GetMessage(),
		},
		resposeType)
	if err!=nil{
		return nil,err
	}
	return &pb.UpdateNotificationEventRes{
		Succeeded: succeeded,
	},nil
}
