package repo

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	connStr += "&default_query_exec_mode=exec"

	testDB, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect to the database via pgxpool: %v", err)
	}

	queries := []string{
		`CREATE TABLE products (
            id BIGSERIAL PRIMARY KEY,
            name varchar(255) NOT NULL,
            image varchar(255) NOT NULL,
            category varchar(255) NOT NULL,
            description text,
            rating int NOT NULL,
            num_reviews int NOT NULL DEFAULT 0,
            price float8 NOT NULL,
            count_in_stock int NOT NULL,
            created_at timestamp DEFAULT now(),
            updated_at timestamp
        );`,
		`CREATE TABLE orders (
            id BIGSERIAL PRIMARY KEY,
            user_id bigint NOT NULL,
            payment_method varchar(255) NOT NULL,
            tax_price float8 NOT NULL,
            shipping_price float8 NOT NULL,
            total_price float8 NOT NULL,
            created_at timestamp DEFAULT now(),
            updated_at timestamp
        );`,
		`CREATE TABLE order_items (
            id BIGSERIAL PRIMARY KEY,
            order_id bigint NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
            product_id bigint NOT NULL REFERENCES products(id) ON DELETE CASCADE,
            name varchar(255) NOT NULL,
            quantity int NOT NULL,
            image varchar(255) NOT NULL,
            price float8 NOT NULL
        );`,
	}

	for _, query := range queries {
		_, err = testDB.Exec(ctx, query)
		if err != nil {
			log.Fatalf("failed to create table: %v", err)
		}
	}

	code := m.Run()

	testDB.Close()
	if err := pgContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func resetDB(t *testing.T) {
	_, err := testDB.Exec(context.Background(), "TRUNCATE TABLE products, orders, order_items RESTART IDENTITY CASCADE")
	require.NoError(t, err, "failed to reset database")
}

func TestCreateProduct(t *testing.T) {
	repo := NewPgxRepo(testDB)

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

				var count int
				err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
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
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestGetProduct(t *testing.T) {
	repo := NewPgxRepo(testDB)

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
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestListProducts(t *testing.T) {
	repo := NewPgxRepo(testDB)

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
				require.Len(t, products, 2)
			},
		},
		{
			name: "empty list",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				products, err := r.ListProducts(ctx)
				require.NoError(t, err)
				require.NotNil(t, products)
				require.Len(t, products, 0)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	repo := NewPgxRepo(testDB)

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
				require.Equal(t, float32(99.99), updated.Price)
			},
		},
		{
			name: "not found",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				p := &Product{ID: 999, Name: "Non-existent", Image: "img", Category: "cat"}
				updated, err := r.UpdateProduct(ctx, p)

				require.Error(t, err)
				require.ErrorIs(t, err, pgx.ErrNoRows)
				require.Nil(t, updated)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	repo := NewPgxRepo(testDB)

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
				err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM products WHERE id=$1", created.ID).Scan(&count)
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
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestCreateOrder(t *testing.T) {
	repo := NewPgxRepo(testDB)

	validOrder := &Order{
		PaymentMethod: "test payment method",
		TaxPrice:      10.0,
		ShippingPrice: 20.0,
		TotalPrice:    129.99,
		Items: []OrderItem{
			{
				Name:      "test product",
				Quantity:  1,
				Image:     "test.jpg",
				Price:     99.99,
				ProductID: 1,
			},
			{
				Name:      "test product 2",
				Quantity:  2,
				Image:     "test2.jpg",
				Price:     199.99,
				ProductID: 2,
			},
		},
	}

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				r.CreateProduct(ctx, &Product{Name: "test product", Price: 99.99, CountInStock: 10})
				r.CreateProduct(ctx, &Product{Name: "test product 2", Price: 199.99, CountInStock: 10})

				created, err := r.CreateOrder(ctx, validOrder)

				require.NoError(t, err)
				require.NotZero(t, created.ID)
				require.Equal(t, validOrder.PaymentMethod, created.PaymentMethod)
				require.Len(t, created.Items, 2)

				// Verify order in the database
				var count int
				err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE id=$1", created.ID).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 1, count)

				// Verify order_items in the database
				err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM order_items WHERE order_id=$1", created.ID).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 2, count)
			},
		},
		{
			name: "database constraint failure (missing payment method)",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()
				invalidOrder := *validOrder
				invalidOrder.PaymentMethod = strings.Repeat("A", 256)

				_, err := r.CreateOrder(ctx, &invalidOrder)
				require.Error(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestGetOrder(t *testing.T) {
	repo := NewPgxRepo(testDB)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				r.CreateProduct(ctx, &Product{Name: "Item 1", Price: 50, CountInStock: 10})
				r.CreateProduct(ctx, &Product{Name: "Item 2", Price: 39.99, CountInStock: 10})

				// Setup data
				o := &Order{
					PaymentMethod: "test payment method",
					TaxPrice:      10.0,
					ShippingPrice: 20.0,
					TotalPrice:    129.99,
					UserID:        1,
					Items: []OrderItem{
						{Name: "Item 1", Quantity: 1, Image: "img1", Price: 50, ProductID: 1},
						{Name: "Item 2", Quantity: 2, Image: "img2", Price: 39.99, ProductID: 2},
					},
				}
				created, err := r.CreateOrder(ctx, o)
				require.NoError(t, err)

				// Test Get
				fetched, err := r.GetOrder(ctx, created.UserID)
				require.NoError(t, err)
				require.NotNil(t, fetched)
				require.Equal(t, created.ID, fetched.ID)
				require.Equal(t, o.PaymentMethod, fetched.PaymentMethod)
				require.Len(t, fetched.Items, 2)
			},
		},
		{
			name: "not found",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				fetched, err := r.GetOrder(ctx, 999) // Non-existent ID
				require.Error(t, err)
				require.ErrorIs(t, err, pgx.ErrNoRows)
				require.Nil(t, fetched)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestListOrders(t *testing.T) {
	repo := NewPgxRepo(testDB)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				r.CreateProduct(ctx, &Product{Name: "1", Price: 10, CountInStock: 10})
				r.CreateProduct(ctx, &Product{Name: "2", Price: 20, CountInStock: 10})

				// Setup data
				o1 := &Order{PaymentMethod: "method 1", TotalPrice: 10, Items: []OrderItem{{Name: "1", Price: 10, ProductID: 1}}}
				o2 := &Order{PaymentMethod: "method 2", TotalPrice: 20, Items: []OrderItem{{Name: "2", Price: 20, ProductID: 2}}}

				_, err := r.CreateOrder(ctx, o1)
				require.NoError(t, err)

				_, err = r.CreateOrder(ctx, o2)
				require.NoError(t, err)

				orders, err := r.ListOrders(ctx)
				require.NoError(t, err)
				require.NotNil(t, orders)
				require.Len(t, orders, 2)
			},
		},
		{
			name: "empty list",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				orders, err := r.ListOrders(ctx)
				require.NoError(t, err)
				require.NotNil(t, orders)
				require.Len(t, orders, 0)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)
			tc.test(t, repo)
		})
	}
}

func TestDeleteOrder(t *testing.T) {
	repo := NewPgxRepo(testDB)

	tcs := []struct {
		name string
		test func(*testing.T, *PgsRepo)
	}{
		{
			name: "success",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				r.CreateProduct(ctx, &Product{Name: "Item 1", Price: 100, CountInStock: 10})

				// Setup data
				o := &Order{
					PaymentMethod: "test method",
					TotalPrice:    100,
					Items: []OrderItem{
						{Name: "Item 1", Price: 100, ProductID: 1},
					},
				}
				created, err := r.CreateOrder(ctx, o)
				require.NoError(t, err)

				// Delete
				err = r.DeleteOrder(ctx, created.ID)
				require.NoError(t, err)

				// Verify Order is deleted
				var count int
				err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE id=$1", created.ID).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 0, count)

				// Verify cascading delete on OrderItems
				err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM order_items WHERE order_id=$1", created.ID).Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			},
		},
		{
			name: "not found",
			test: func(t *testing.T, r *PgsRepo) {
				ctx := context.Background()

				err := r.DeleteOrder(ctx, 999)
				require.Error(t, err)
				require.ErrorIs(t, err, pgx.ErrNoRows)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			resetDB(t)
			tc.test(t, repo)
		})
	}
}
