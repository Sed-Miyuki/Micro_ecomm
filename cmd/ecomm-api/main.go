package main

import (
	"log"
	"os"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/handler"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/pb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const minSecretKeySize=32

func main(){
	if err:=godotenv.Load();err!=nil{
		_=godotenv.Load("../../.env")
	}
	secretKey:=os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("secret_key is not set in environment variables")
	}
	if len(secretKey)<minSecretKeySize{
		log.Fatalf("secret_key must be atleast %d characters",minSecretKeySize)
	}

	scvAddr:=os.Getenv("SCV_ADDR")
	if scvAddr == "" {
		log.Fatal("Service address is not set in environment variables")
	}

	opts:=[]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn,err:=grpc.NewClient(scvAddr,opts...)
	if err!=nil{
		log.Fatalf("failed to connect to server: %v",err)
	}
	defer conn.Close()

	client:=pb.NewEcommClient(conn)

	hdl:=handler.NewHandler(client,secretKey)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")
}