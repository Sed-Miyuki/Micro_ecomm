package main

import (
	"log"

	"github.com/Sed-Miyuki/Micro_ecomm/db"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/handler"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/repo"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/server"
)

func main(){
	db,err:=db.NewDatabase()
	if err!=nil{
		log.Fatalf("error opening database: %v",err)
	}
	defer db.Close()
	log.Println("successfully connected to database")

	pgs:=repo.NewPgxRepo(db.GetDB())
	srv:=server.NewServer(pgs)
	hdl:=handler.NewHandler(srv)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")
}