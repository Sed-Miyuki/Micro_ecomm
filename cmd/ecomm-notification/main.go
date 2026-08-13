package main

import (
	"context"
	"log"
	"os"

	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-grpc/pb"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-notification/server"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main(){
	if err:=godotenv.Load();err!=nil{
		_=godotenv.Load("../../.env")
	}
	svcAddr:=os.Getenv("SCV_ADDR")
	if svcAddr == "" {
		log.Fatal("SVC_ADDR is not set in environment variables")
	}
	adminEmail:=os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		log.Fatal("ADMIN_EMAIL is not set in environment variables")
	}
	adminPassword:=os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatal("ADMIN_PASSWORD is not set in environment variables")
	}

	opts:=[]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn,err:=grpc.NewClient(svcAddr,opts...)
	if err!=nil{
		log.Fatalf("failed to connect to server: %v",err)
	}
	defer conn.Close()

	client:=pb.NewEcommClient(conn)
	srv:=server.NewServer(client,&server.AdminInfo{
		Email: adminEmail,
		Password: adminPassword,
	})

	done:=make(chan struct{})
	go func ()  {
		srv.RUN(context.Background())
		done<-struct{}{}
	}()
	<-done
}
