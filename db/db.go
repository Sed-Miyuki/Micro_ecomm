package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Database struct{
	db *pgxpool.Pool
}

func NewDatabase() (*Database,error){
	if err:=godotenv.Load();err!=nil{
		_=godotenv.Load("../../.env")
	}
	dbURL:=os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set in environment variables")
	}
	ctx:=context.Background()
	connStr:=dbURL
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
