package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		return nil, err
	}
	defer res.Close() 

	p, err = pgx.CollectOneRow(res, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (pgs *PgsRepo) ListProducts(ctx context.Context) ([]Product, error) {
	query := "SELECT id, name, image, category, description, rating, num_reviews, price, count_in_stock, created_at, updated_at FROM products"
	res, err := pgs.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	p, err := pgx.CollectRows(res, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (pgs *PgsRepo) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	query := `UPDATE products 
	          SET name = $1, image = $2, category = $3, description = $4, 
	              rating = $5, num_reviews = $6, price = $7, count_in_stock = $8, 
	              updated_at = $9
	          WHERE id = $10
	          RETURNING id, name, image, category, description, rating, 
	                    num_reviews, price, count_in_stock, created_at, updated_at`
	rows, err := pgs.db.Query(ctx, query,
		p.Name, p.Image, p.Category, p.Description,
		p.Rating, p.NumReviews, p.Price, p.CountInStock,
		p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() 

	updatedProduct, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, err
	}
	return &updatedProduct, nil
}

func (pgs *PgsRepo) DeleteProduct(ctx context.Context, id int64) error {
	query := "DELETE FROM products WHERE id=$1"
	res, err := pgs.db.Exec(ctx, query, id)
	if err != nil {
		return err
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
			o.Items[i].OrderId = o.ID
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
	query := "INSERT INTO orders(payment_method,tax_price,shipping_price,total_price) VALUES($1,$2,$3,$4) RETURNING id, created_at, updated_at;"
	err := tx.QueryRow(ctx, query, o.PaymentMethod, o.TaxPrice, o.ShippingPrice, o.TotalPrice).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}
	return o, err
}

func createOrderItem(ctx context.Context, tx pgx.Tx, oi *OrderItem) (*OrderItem, error) {
	query := "INSERT INTO order_items(name,quantity,image,price,product_id,order_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING id;"
	err := tx.QueryRow(ctx, query, oi.Name, oi.Quantity, oi.Image, oi.Price, oi.ProductId, oi.OrderId).Scan(&oi.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order_item: %w", err)
	}
	return oi, nil
}

func (pgs *PgsRepo) GetOrder(ctx context.Context, id int64) (*Order, error) {
	query := "SELECT id, payment_method, tax_price, shipping_price, total_price, created_at, updated_at FROM orders WHERE id=$1"
	res, err := pgs.db.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting order: %w", err)
	}
	defer res.Close() // 💡 Connection leak patched

	o, err := pgx.CollectOneRow(res, pgx.RowToStructByName[Order])
	if err != nil {
		return nil, fmt.Errorf("error collecting order row: %w", err)
	}

	query = "SELECT id, name, quantity, image, price, product_id, order_id FROM order_items WHERE order_id=$1"
	resi, err := pgs.db.Query(ctx, query, id)
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

func (pgs *PgsRepo) ListOrders(ctx context.Context) ([]Order, error) {
	query := "SELECT id, payment_method, tax_price, shipping_price, total_price, created_at, updated_at FROM orders"
	res, err := pgs.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting orders: %w", err)
	}
	defer res.Close()

	orders, err := pgx.CollectRows(res, pgx.RowToStructByName[Order])
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
		orderMap[orders[i].ID] = &orders[i]
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
		if parentOrder, exists := orderMap[item.OrderId]; exists {
			parentOrder.Items = append(parentOrder.Items, item)
		}
	}
	return orders, nil
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
