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
	query := "INSERT INTO products (name,image,category,description,rating,num_reviews,price,count_in_stock) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"
	err := pgs.db.QueryRow(ctx, query, p.Name, p.Image, p.Category, p.Description, p.Rating, p.NumReviews, p.Price, p.CountInStock).Scan(&p.ID)
	if err != nil {
		return nil, fmt.Errorf("error inserting product: %v", err)
	}
	return p, nil
}

func (pgs *PgsRepo) GetProduct(ctx context.Context, id int64) (*Product, error) {
	var p Product
	query := "SELECT * FROM products WHERE id=$1"
	res, err := pgs.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	p, err = pgx.CollectOneRow(res, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, err
	}
	return &p, err
}

func (pgs *PgsRepo) ListProducts(ctx context.Context) (*[]Product, error) {
	var p []Product
	query := "SELECT * FROM products"
	res, err := pgs.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	p, err = pgx.CollectRows(res, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, err
	}
	return &p, err
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
		return nil, err
	}
	updatedProduct, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
	if err != nil {
		return nil, err
	}
	return &updatedProduct, nil
}

func (pgs *PgsRepo) DeleteProduct(ctx context.Context,id int64) error {
	query := "DELETE FROM products WHERE id=$1"
	res, err := pgs.db.Exec(ctx, query,id)
	if err != nil {
		return err
	}
	if res.RowsAffected()==0 {
		return pgx.ErrNoRows
	}
	return nil
}