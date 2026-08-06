package main

import (
	"log"
	"os"

	"github.com/Sed-Miyuki/Micro_ecomm/db"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/handler"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/repo"
	"github.com/Sed-Miyuki/Micro_ecomm/ecomm-api/server"
	"github.com/joho/godotenv"
)

const minSecretKeySize=32

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
	secretKey:=os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("secret_key is not set in environment variables")
	}
	if len(secretKey)<minSecretKeySize{
		log.Fatalf("secret_key must be atleast %d characters",minSecretKeySize)
	}

	pgs:=repo.NewPgxRepo(db.GetDB())
	srv:=server.NewServer(pgs)
	hdl:=handler.NewHandler(srv,secretKey)
	handler.RegisterRoutes(hdl)
	handler.Start(":8080")
}