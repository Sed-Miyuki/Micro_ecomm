package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string,error){
	//cost is how many times you want to hash
	hashed,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err!=nil{
		return "",fmt.Errorf("error hashing password: %w",err)
	}
	return string(hashed),nil
}

func CheckPassword(password string,hashedPassword string) error{
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte(password))
}