package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct{
	secretkey string
}

func NewJWTMaker(secretkey string) *JWTMaker{
	return &JWTMaker{secretkey: secretkey}
}

func (maker *JWTMaker) CreateToken(id int64,email string,isAdmin bool,duration time.Duration) (string,*UserClaims,error){
	claims,err:=NewUserClaims(id,email,isAdmin,duration)
	if err!=nil{
		return "",nil,fmt.Errorf("error making claims: %w",err)
	}
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenstr,err:=token.SignedString([]byte(maker.secretkey))
	if err!=nil{
		return "",nil,fmt.Errorf("error signing token: %w",err)
	}
	return tokenstr,claims,nil
}

func (maker *JWTMaker) VerifyToken(tokenstr string) (*UserClaims,error){
	token,err:=jwt.ParseWithClaims(tokenstr,&UserClaims{},func(t *jwt.Token) (interface{}, error) {
		_,ok:=t.Method.(*jwt.SigningMethodHMAC)
		if !ok{
			return nil,fmt.Errorf("invalid token signing method")
		}
		return []byte(maker.secretkey),nil
	})
	if err!=nil{
		return nil,fmt.Errorf("error parsing token: %w",err)
	}
	claims,ok:=token.Claims.(*UserClaims)
	if !ok{
		return nil,fmt.Errorf("invalid token claims: %w",err)
	}
	return claims,nil
}