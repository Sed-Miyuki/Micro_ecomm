package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const(
	maxAttempts=3
)

type PgsRepo struct {
	db *pgxpool.Pool
}

func NewPgxRepo(db *pgxpool.Pool) *PgsRepo {
	return &PgsRepo{db: db}
}

func (pgs *PgsRepo) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	query := "INSERT INTO products (name,image,category,description,rating,num_reviews,price,count_in_stock) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at, updated_at"
	err := pgs.db.QueryRow(ctx, query, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting product: %w", err)
	}
	return p, nil
}

func (pgs *PgsRepo) GetProduct(ctx context.Context, id int64) (*Product, error) {
	var p Product
	query := `SELECT id, name, image, category, description, rating, 
                     num_reviews, price, count_in_stock, created_at, updated_at 
              FROM products WHERE id=$1`
	res, err := pgs.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting product: %w", err)
	}
	defer res.Close()

	p, err = pgx.CollectOneRow(res, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, fmt.Errorf("error getting product row: %w", err)
	}
	return &p, nil
}

func (pgs *PgsRepo) ListProducts(ctx context.Context) ([]*Product, error) {
	query := "SELECT id, name, image, category, description, rating, num_reviews, price, count_in_stock, created_at, updated_at FROM products"
	res, err := pgs.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting products: %w", err)
	}
	defer res.Close()

	p, err := pgx.CollectRows(res, pgx.RowToAddrOfStructByName[Product])
	if err != nil {
		return nil, fmt.Errorf("error getting products rows: %w", err)
	}
	return p, nil
}

func (pgs *PgsRepo) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	query := `UPDATE products 
              SET name = $1, image = $2, category = $3, description = $4, 
                  rating = $5, num_reviews = $6, price = $7, count_in_stock = $8, 
                  updated_at = NOW()
              WHERE id = $9
              RETURNING id, name, image, category, description, rating, 
                        num_reviews, price, count_in_stock, created_at, updated_at`
	rows, err := pgs.db.Query(ctx, query,
		p.Name, p.Image, p.Category, p.Description,
		p.Rating, p.NumReviews, p.Price, p.CountInStock,
		p.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("error updating product: %w", err)
	}
	defer rows.Close()

	updatedProduct, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, fmt.Errorf("error updating product row: %w", err)
	}
	return &updatedProduct, nil
}

func (pgs *PgsRepo) DeleteProduct(ctx context.Context, id int64) error {
	query := "DELETE FROM products WHERE id=$1"
	res, err := pgs.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (pgs *PgsRepo) CreateOrder(ctx context.Context, o *Order) (*Order, error) {
	err := pgs.execTx(ctx, func(tx pgx.Tx) error {
		var err error
		o, err = createOrder(ctx, tx, o)
		if err != nil {
			return err
		}
		for i := range o.Items {
			o.Items[i].OrderID = o.ID
			_, err := createOrderItem(ctx, tx, &o.Items[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}
	return o, nil
}

func createOrder(ctx context.Context, tx pgx.Tx, o *Order) (*Order, error) {
	query := "INSERT INTO orders(payment_method,tax_price,shipping_price,total_price,user_id) VALUES($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at;"
	err := tx.QueryRow(ctx, query, o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice, o.UserID).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}
	return o, err
}

func createOrderItem(ctx context.Context, tx pgx.Tx, oi *OrderItem) (*OrderItem, error) {
	query := "INSERT INTO order_items(name,quantity,image,price,product_id,order_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING id;"
	err := tx.QueryRow(ctx, query, oi.Name, oi.Quantity, oi.Image, oi.Price, oi.ProductID, oi.OrderID).Scan(&oi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order_item: %w", err)
	}
	return oi, nil
}

func (pgs *PgsRepo) GetOrder(ctx context.Context, userID int64) (*Order, error) {
	query := "SELECT id, user_id, payment_method, tax_price, shipping_price, total_price, created_at, updated_at, status FROM orders WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1"
	res, err := pgs.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting order: %w", err)
	}
	defer res.Close()

	o, err := pgx.CollectOneRow(res, pgx.RowToStructByName[Order])
	if err != nil {
		return nil, fmt.Errorf("error collecting order row: %w", err)
	}

	query = "SELECT id, name, quantity, image, price, product_id, order_id FROM order_items WHERE order_id=$1"
	resi, err := pgs.db.Query(ctx, query, o.ID)
	if err != nil {
		return nil, fmt.Errorf("error querying order items: %w", err)
	}
	defer resi.Close()

	items, err := pgx.CollectRows(resi, pgx.RowToStructByName[OrderItem])
	if err != nil {
		return nil, fmt.Errorf("error collecting order_item rows: %w", err)
	}
	o.Items = items
	return &o, nil
}

func (pgs *PgsRepo) GetOrderStatusByID(ctx context.Context,id int64) (*Order,error){
	query := "SELECT id, user_id, payment_method, tax_price, shipping_price, total_price, created_at, updated_at, status FROM orders WHERE id=$1 ORDER BY created_at DESC LIMIT 1"
	res, err := pgs.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting order: %w", err)
	}
	defer res.Close()

	o, err := pgx.CollectOneRow(res, pgx.RowToStructByName[Order])
	if err != nil {
		return nil, fmt.Errorf("error collecting order row: %w", err)
	}

	query = "SELECT id, name, quantity, image, price, product_id, order_id FROM order_items WHERE order_id=$1"
	resi, err := pgs.db.Query(ctx, query, o.ID)
	if err != nil {
		return nil, fmt.Errorf("error querying order items: %w", err)
	}
	defer resi.Close()

	items, err := pgx.CollectRows(resi, pgx.RowToStructByName[OrderItem])
	if err != nil {
		return nil, fmt.Errorf("error collecting order_item rows: %w", err)
	}
	o.Items = items
	return &o, nil
}

func (pgs *PgsRepo) ListOrders(ctx context.Context) ([]*Order, error) {
	query := "SELECT id,user_id, payment_method, tax_price, shipping_price, total_price,status, created_at, updated_at FROM orders"
	res, err := pgs.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting orders: %w", err)
	}
	defer res.Close()

	orders, err := pgx.CollectRows(res, pgx.RowToAddrOfStructByName[Order])
	if err != nil {
		return nil, fmt.Errorf("error collecting orders: %w", err)
	}
	if len(orders) == 0 {
		return orders, nil
	}

	orderIDs := make([]int64, len(orders))
	orderMap := make(map[int64]*Order, len(orders))
	for i := range orders {
		orderIDs[i] = orders[i].ID
		orderMap[orders[i].ID] = orders[i]
	}

	itemQuery := "SELECT id, name, quantity, image, price, product_id, order_id FROM order_items WHERE order_id=ANY($1)"
	resi, err := pgs.db.Query(ctx, itemQuery, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("error querying order items: %w", err)
	}
	defer resi.Close()

	items, err := pgx.CollectRows(resi, pgx.RowToStructByName[OrderItem])
	if err != nil {
		return nil, fmt.Errorf("error collecting order items: %w", err)
	}

	for _, item := range items {
		if parentOrder, exists := orderMap[item.OrderID]; exists {
			parentOrder.Items = append(parentOrder.Items, item)
		}
	}
	return orders, nil
}

func (pgs *PgsRepo) UpdateOrderStatus(ctx context.Context,o *Order) (*Order,error){
	query := `
		UPDATE orders 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
		RETURNING id, user_id, payment_method, tax_price, shipping_price, total_price, status, created_at, updated_at`
	
	rows, err := pgs.db.Query(ctx, query, o.Status, o.ID)
	if err != nil {
		return nil, fmt.Errorf("error updating order status: %w", err)
	}
	defer rows.Close()
	updatedOrder, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Order])
	if err != nil {
		return nil, fmt.Errorf("error collecting updated order: %w", err)
	}
	return &updatedOrder, nil
}

func (pgs *PgsRepo) DeleteOrder(ctx context.Context, id int64) error {
	return pgs.execTx(ctx, func(tx pgx.Tx) error {
		query := "DELETE FROM order_items WHERE order_id=$1"
		_, err := tx.Exec(ctx, query, id)
		if err != nil {
			return fmt.Errorf("error deleting order items: %w", err)
		}

		query = "DELETE FROM orders WHERE id=$1"
		tag, err := tx.Exec(ctx, query, id)
		if err != nil {
			return fmt.Errorf("error deleting orders: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return nil
	})
}

func (pgs *PgsRepo) CreateUser(ctx context.Context, u *User) (*User, error) {
	query := "INSERT INTO users (name,email,password,is_admin) VALUES($1,$2,$3,$4) RETURNING id,created_at,updated_at"
	err := pgs.db.QueryRow(ctx, query, u.Name, u.Email, u.Password, u.IsAdmin).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return u, nil
}

func (pgs *PgsRepo) GetUser(ctx context.Context, email string) (*User, error) {
	var u User
	query := "SELECT id,name,email,password,is_admin,created_at,updated_at FROM users WHERE email=$1"
	res, err := pgs.db.Query(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	defer res.Close()
	u, err = pgx.CollectOneRow(res, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("error collecting user row: %w", err)
	}
	return &u, nil
}

func (pgs *PgsRepo) ListUsers(ctx context.Context) ([]*User, error) {
	query := "SELECT id,name,email,password,is_admin,created_at,updated_at FROM users"
	res, err := pgs.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer res.Close()
	u, err := pgx.CollectRows(res, pgx.RowToAddrOfStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("error collecting users row: %w", err)
	}
	return u, nil
}

func (pgs *PgsRepo) UpdateUser(ctx context.Context, u *User) (*User, error) {
	query := "UPDATE users SET name=$1,email=$2,password=$3,is_admin=$4,updated_at=NOW() WHERE id=$5 RETURNING id,name,email,password,is_admin,created_at,updated_at"
	res, err := pgs.db.Query(ctx, query, u.Name, u.Email, u.Password, u.IsAdmin, u.ID)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}
	defer res.Close()
	updated_user, err := pgx.CollectOneRow(res, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("error updating user row: %w", err)
	}
	return &updated_user, nil
}

func (pgs *PgsRepo) DeleteUser(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id=$1"
	res, err := pgs.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (pgs *PgsRepo) CreateSession(ctx context.Context, s *Session) (*Session, error) {
	query := "INSERT INTO sessions (id,user_email,refresh_token,is_revoked,expires_at) VALUES ($1,$2,$3,$4,$5) RETURNING created_at"
	err := pgs.db.QueryRow(ctx, query, s.ID, s.UserEmail, s.RefreshToken, s.IsRevoked, s.ExpiresAt).Scan(&s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}
	return s, nil
}

func (pgs *PgsRepo) GetSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	query := "SELECT id,user_email,is_revoked,refresh_token,expires_at,created_at FROM sessions WHERE id=$1"
	res, err := pgs.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting session: %w", err)
	}
	defer res.Close()
	s, err = pgx.CollectOneRow(res, pgx.RowToStructByName[Session])
	if err != nil {
		return nil, fmt.Errorf("error getting session row: %w", err)
	}
	return &s, nil
}

func (pgs *PgsRepo) RevokeSession(ctx context.Context, id string) error {
	query := "UPDATE sessions SET is_revoked=TRUE WHERE id=$1"
	_, err := pgs.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error revoking session %w", err)
	}
	return nil
}

func (pgs *PgsRepo) DeleteSession(ctx context.Context, id string) error {
	query := "DELETE FROM sessions WHERE id=$1"
	_, err := pgs.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting session :%w", err)
	}
	return nil
}

func (pgs *PgsRepo) execTx(ctx context.Context, fn func(pgx.Tx) error) (err error) {
	tx, txErr := pgs.db.Begin(ctx)
	if txErr != nil {
		return fmt.Errorf("error starting transaction: %w", txErr)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	err = fn(tx)
	if err != nil {
		return fmt.Errorf("error in transaction execution: %w", err)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("error committing transaction: %w", commitErr)
	}
	return nil
}

func insertNotificationState(ctx context.Context,tx pgx.Tx,es *NotificationState) (*NotificationState,error){
	query:="INSERT INTO notification_states (order_id,state,message) VALUES ($1,$2,$3) RETURNING id,requested_at,completed_at"
	err:=tx.QueryRow(ctx,query,es.OrderID,es.State,es.Message).Scan(&es.ID,&es.RequestedAt,&es.CompletedAt)
	if err!=nil{
		return nil,fmt.Errorf("error inserting into notification table: %w",err)
	}
	return es,nil
}

func insertNotificationEvent(ctx context.Context,tx pgx.Tx,u *NotificationEvent) (*NotificationEvent,error){
	query:="INSERT INTO notification_events_queue (user_email,order_status,order_id,state_id,attempts) VALUES ($1,$2,$3,$4,$5) RETURNING id,created_at,updated_at"
	err:=tx.QueryRow(ctx,query,u.UserEmail,u.OrderStatus,u.OrderID,u.StateID,u.Attempts).Scan(&u.ID,&u.CreatedAt,&u.UpdatedAt)
	if err!=nil{
		return nil,fmt.Errorf("error inserting into notification_events_queue table: %w",err)
	}
	return u,nil
}

func (pgs *PgsRepo) EnqueueNotificationEvent(ctx context.Context,ne *NotificationEvent) (*NotificationEvent,error){
	var ev *NotificationEvent
	err:=pgs.execTx(ctx,func(tx pgx.Tx) error {
		ns,err:=insertNotificationState(ctx,tx,&NotificationState{
			OrderID: ne.OrderID,
			State: NotSent,
			Message: "",
		})
		if err!=nil{
			return fmt.Errorf("error inserting notification state: %w",err)
		}
		ne.StateID=ns.ID
		ev,err=insertNotificationEvent(ctx,tx,ne)
		if err!=nil{
			return fmt.Errorf("error inserting notification event: %w",err)
		}
		return nil
	})
	if err!=nil{
		return nil,fmt.Errorf("error enqueuing notification event: %w",err)
	}
	return ev,nil
}

func (pgs *PgsRepo) ListNotificationEvents(ctx context.Context) ([]*NotificationEvent,error){
	var events []*NotificationEvent
	query:="SELECT id,user_email,order_status,order_id,state_id,attempts,created_at,updated_at FROM notification_events_queue WHERE attempts<$1 ORDER BY created_at"
	res,err:=pgs.db.Query(ctx,query,maxAttempts)
	if err!=nil{
		return nil,fmt.Errorf("error getting notification events: %w",err)
	}
	rows,err:=pgx.CollectRows(res,pgx.RowToAddrOfStructByName[NotificationEvent])
	if err!=nil{
		return nil,fmt.Errorf("error getting notification events rows: %w",err)
	}
	for _,i:=range rows{
		events = append(events, i)
	}
	return events,nil
}

func getNotificationEventAttempts(ctx context.Context,tx pgx.Tx,id int64) (*NotificationEvent,error){
	query:="SELECT id,attempts FROM notification_events_queue WHERE id=$1"
	res,err:=tx.Query(ctx,query,id)
	if err!=nil{
		return nil,fmt.Errorf("error getting notification event: %w",err)
	}
	ne,err:=pgx.CollectOneRow(res,pgx.RowToAddrOfStructByNameLax[NotificationEvent])
	if err!=nil{
		return nil,fmt.Errorf("error getting notification event row: %w",err)
	}
	return ne,nil
}

func updateNotificationEventAttempts(ctx context.Context,tx pgx.Tx,u *NotificationEvent) (*NotificationEvent,error){
	query:="UPDATE notification_events_queue SET attempts=$1,updated_at=NOW() WHERE id=$2 RETURNING updated_at"
	err:=tx.QueryRow(ctx,query,u.Attempts,u.ID).Scan(&u.UpdatedAt)
	if err!=nil{
		return nil,fmt.Errorf("error updating notification event attempt: %w",err)
	}
	return u,nil
}

func deleteNotificationEvent(ctx context.Context,tx pgx.Tx,id int64) error{
	query:="DELETE FROM notification_events_queue WHERE id=$1"
	_,err:=tx.Exec(ctx,query,id)
	if err!=nil{
		return fmt.Errorf("error deleting notification event: %w",err)
	}
	return nil
}

func updateNotificationState(ctx context.Context,tx pgx.Tx,es *NotificationState) error{
	if es.State==Sent{
		t:=time.Now()
		es.CompletedAt=&t
		query:="UPDATE notification_states SET state=$1,message=$2,completed_at=$3 WHERE id=$4"
		_,err:=tx.Exec(ctx,query,es.State,es.Message,es.CompletedAt,es.ID)
		if err!=nil{
			return fmt.Errorf("error updating notification")
		}
		return nil
	}else{
		query:="UPDATE notification_states SET state=$1,message=$2 WHERE id=$3"
		_,err:=tx.Exec(ctx,query,es.State,es.Message,es.ID)
		if err!=nil{
			return fmt.Errorf("error updating notification")
		}
		return nil
	}
}

func (pgs *PgsRepo) UpdateNotificationEvent(ctx context.Context,ev *NotificationEvent,es *NotificationState,responseType NotificationResponseType) (bool,error){
	succeeded:=false
	err:=pgs.execTx(ctx,func(tx pgx.Tx) error {
		switch responseType{
		case NotificationSucess:
			err:=updateNotificationState(ctx,tx,&NotificationState{
				ID: ev.StateID,
				State: Sent,
				Message: es.Message,
			})
			if err!=nil{
				return fmt.Errorf("error updating notification state: %w",err)
			}
			err=deleteNotificationEvent(ctx,tx,ev.ID)
			if err!=nil{
				return fmt.Errorf("error deleting notification event: %w",err)
			}
			succeeded=true
		case NotificationFailure:
			u,err:=getNotificationEventAttempts(ctx,tx,ev.ID)
			if err!=nil{
				return fmt.Errorf("error getting notification event: %w",err)
			}
			if u.Attempts+1<maxAttempts{
				t:=time.Now()
				u.UpdatedAt=&t
				u.Attempts+=1

				_,err:=updateNotificationEventAttempts(ctx,tx,u)
				if err!=nil{
					return fmt.Errorf("error updating notification event: %w",err)
				}
			}else{
				err=updateNotificationState(ctx,tx,&NotificationState{
					ID: ev.StateID,
					State: Failed,
					Message: es.Message,
				})
				if err!=nil{
					return fmt.Errorf("error updating notification state: %w",err)
				}
				err=deleteNotificationEvent(ctx,tx,u.ID)
				if err!=nil{
					return fmt.Errorf("error deleting notification event: %w",err)
				}
			}
		default:
			return fmt.Errorf("invalid notification response type: %v",responseType)
		}
		return nil
	})
	return succeeded,err
}