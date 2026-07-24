package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct{
	db *pgxpool.Pool
}

func NewDatabase() (*Database,error){
	ctx:=context.Background()
	connStr:="postgres://postgres:password@localhost:5433/ecomm?sslmode=disable"
	db,err:=pgxpool.New(ctx,connStr)
	if err!=nil{
		return nil,fmt.Errorf("error opening database: %v",err)
	}
	return &Database{db:db},nil
}

func (d *Database) Close() error{
	d.db.Close()
	return nil
}

func (d *Database) GetDB() *pgxpool.Pool{
	return d.db
}
