package repo

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// setupTestDB spins up a PostgreSQL container, creates the connection pool,
// sets up the schema, and returns the pool. It registers cleanups automatically.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err, "failed to start container")

	t.Cleanup(func() {
		err := pgContainer.Terminate(ctx)
		require.NoError(t, err, "failed to terminate container")
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "failed to connect to the database via pgxpool")

	t.Cleanup(func() {
		db.Close()
	})

	_, err = db.Exec(ctx, `
        CREATE TABLE products (
            id SERIAL PRIMARY KEY,
            name varchar(255) NOT NULL,
            image varchar(255) NOT NULL,
            category varchar(255) NOT NULL,
            description text,
            rating int NOT NULL,
            num_reviews int NOT NULL DEFAULT 0,
            price decimal(10,2) NOT NULL,
            count_in_stock int NOT NULL,
            created_at timestamp DEFAULT now(),
            updated_at timestamp
        )
    `)
	require.NoError(t, err, "failed to create table")

	return db
}

// resetDB clears the table between test cases to ensure isolation
func resetDB(t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(context.Background(), "TRUNCATE TABLE products RESTART IDENTITY CASCADE")
	require.NoError(t, err, "failed to reset database")
}

func TestCreateProduct(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPgxRepo(db)

	validProduct := &Product{
		Name:         "test_product",
		Image:        "test_image",
		Category:     "test_category",
		Description:  "test_description",
		Rating:       5,
		NumReviews:   10,
		Price:        100.0,
		CountInStock: 100,
	}

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				created, err := r.CreateProduct(ctx, validProduct)
				
				require.NoError(t, err)
				require.NotZero(t, created.ID)
				require.Equal(t, validProduct.Name, created.Name)

				// Verify it's actually in the database
				var count int
				err = db.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 1, count)
			},
		},
		{
			name: "database constraint failure (missing name)",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				invalidProduct := *validProduct
				invalidProduct.Name = strings.Repeat("A", 256)
				_, err := r.CreateProduct(ctx, &invalidProduct)
				require.Error(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, db)
			tc.test(t, repo)
		})
	}
}

func TestGetProduct(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPgxRepo(db)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				// Setup data
				p := &Product{Name: "P1", Image: "img", Category: "cat", Rating: 5, Price: 50, CountInStock: 10}
				created, err := r.CreateProduct(ctx, p)
				require.NoError(t, err)

				// Test Get
				fetched, err := r.GetProduct(ctx, created.ID)
				require.NoError(t, err)
				require.Equal(t, created.ID, fetched.ID)
				require.Equal(t, "P1", fetched.Name)
			},
		},
		{
			name: "not found",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				fetched, err := r.GetProduct(ctx, 999) // Non-existent ID
				require.Error(t, err)
				require.ErrorIs(t, err, pgx.ErrNoRows)
				require.Nil(t, fetched)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, db)
			tc.test(t, repo)
		})
	}
}

func TestListProducts(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPgxRepo(db)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				// Setup data
				r.CreateProduct(ctx, &Product{Name: "P1", Image: "img", Category: "cat", Rating: 5, Price: 50, CountInStock: 10})
				r.CreateProduct(ctx, &Product{Name: "P2", Image: "img", Category: "cat", Rating: 4, Price: 60, CountInStock: 20})

				products, err := r.ListProducts(ctx)
				require.NoError(t, err)
				require.NotNil(t, products)
				require.Len(t, *products, 2)
			},
		},
		{
			name: "empty list",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				products, err := r.ListProducts(ctx)
				require.NoError(t, err)
				require.NotNil(t, products)
				require.Len(t, *products, 0)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, db)
			tc.test(t, repo)
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPgxRepo(db)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				// Setup data
				p := &Product{Name: "Old", Image: "img", Category: "cat", Rating: 5, Price: 50, CountInStock: 11}
				created, err := r.CreateProduct(ctx, p)
				require.NoError(t, err)

				// Update data
				created.Name = "Updated"
				created.Price = 99.99

				updated, err := r.UpdateProduct(ctx, created)
				require.NoError(t, err)
				require.NotNil(t, updated)
				require.Equal(t, "Updated", updated.Name)
				require.Equal(t, float64(99.99), updated.Price)
			},
		},
		{
			name: "not found",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				p := &Product{ID: 999, Name: "Non-existent", Image: "img", Category: "cat"}
				updated, err := r.UpdateProduct(ctx, p)
				
				require.Error(t, err)
				require.ErrorIs(t, err, pgx.ErrNoRows) // CollectOneRow returns this on empty result set
				require.Nil(t, updated)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, db)
			tc.test(t, repo)
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPgxRepo(db)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				// Setup data
				p := &Product{Name: "To Delete", Image: "img", Category: "cat", Rating: 5, Price: 50, CountInStock: 10}
				created, err := r.CreateProduct(ctx, p)
				require.NoError(t, err)

				// Delete
				err = r.DeleteProduct(ctx, created.ID)
				require.NoError(t, err)

				// Verify
				var count int
				err = db.QueryRow(ctx, "SELECT COUNT(*) FROM products WHERE id=$1", created.ID).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			},
		},
		{
			name: "not found",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				
				err := r.DeleteProduct(ctx, 999)
				require.Error(t, err)
				require.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t, db)
			tc.test(t, repo)
		})
	}
}