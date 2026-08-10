package main

import (
	"log"
	"net"
	"os"

	"github.com/Sed-Miyuki/Micro_ecomm/db"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/pb"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/repo"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/server"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main(){
	db,err:=db.NewDatabase()
	if err!=nil{
		log.Fatalf("error opening database: %v",err)
	}
	defer db.Close()
	log.Println("successfully connected to database")

	if err:=godotenv.Load();err!=nil{
		_=godotenv.Load("../../.env")
	}

	pgs:=repo.NewPgxRepo(db.GetDB())
	srv:=server.NewServer(pgs)

	grpcSrv:=grpc.NewServer()
	pb.RegisterEcommServer(grpcSrv,srv)

	scvAddr:=os.Getenv("SCV_ADDR")
	if scvAddr == "" {
		log.Fatal("Service address is not set in environment variables")
	}
	listener,err:=net.Listen("tcp",scvAddr)
	if err!=nil{
		log.Fatalf("listener failed: %v",err)
	}
	log.Printf("server listening on: %s",scvAddr)
	err=grpcSrv.Serve(listener)
	if err!=nil{
		log.Fatalf("failed to serve: %v",err)
	}
}
