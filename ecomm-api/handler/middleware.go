package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Sed-Miyuki/Micro_ecomm/token"
)

//empty struct ->no memory allocation,faster and better comparision without traversing string
type authkey struct{}
func GetAuthMiddlewareFunc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler{
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims,err:=verifyClaimsFromAuthHeader(r,tokenMaker)
			if err!=nil{
				http.Error(w,fmt.Sprintf("error verifying token: %v",err),http.StatusUnauthorized)
				return
			}
			ctx:=context.WithValue(r.Context(),authkey{},claims)
			next.ServeHTTP(w,r.WithContext(ctx))
		})
	}
}
func GetAdminMiddlewareFunc(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler{
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims,err:=verifyClaimsFromAuthHeader(r,tokenMaker)
			if err!=nil{
				http.Error(w,fmt.Sprintf("error verifying token: %v",err),http.StatusUnauthorized)
				return
			}
			if !claims.IsAdmin{
				http.Error(w,"user is not an admin",http.StatusForbidden)
				return
			}
			ctx:=context.WithValue(r.Context(),authkey{},claims)
			next.ServeHTTP(w,r.WithContext(ctx))
		})
	}
}

func verifyClaimsFromAuthHeader(r *http.Request,tokenmaker *token.JWTMaker) (*token.UserClaims,error){
	authHeader:=r.Header.Get("Authorization")
	if authHeader==""{
		return nil,fmt.Errorf("authorization header missing")
	}
	fields:=strings.Fields(authHeader)
	if len(fields)!=2 || fields[0]!="Bearer"{
		return nil,fmt.Errorf("invalid authorization header ")
	}
	token:=fields[1]
	//VerifyToken.ParseWithClaims automatically checks expiration etc
	claims,err:=tokenmaker.VerifyToken(token)
	if err!=nil{
		return nil,fmt.Errorf("invalid token: %w",err)
	}
	return claims,nil
}