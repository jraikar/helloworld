package auth

import (
	"context"
	"fmt"
	//"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/gorilla/handlers"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"net/http"
	"os"
	"strings"
)

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//verify if request has auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "no token passed in header", http.StatusBadRequest)
			return
		}

		splitAuthHeader := strings.Split(authHeader, " ")
		if len(splitAuthHeader) != 2 {
			http.Error(w, "invalid header", http.StatusBadRequest)
			return
		}

		cognitoClient, err := Init()
		if err != nil {
			http.Error(w, "unable to initialize cognito client", http.StatusInternalServerError)
			return
		}

		publicKeyURL := "https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json"
		formattedURLString := fmt.Sprintf(publicKeyURL, os.Getenv("AWS_REGION"), cognitoClient.UserPoolId)
		//todo: the publicKeyURL is static and it's not the recommend way as per best security practices, we should use jwk.NewAutoRefresh()

		// Use jwk.AutoRefresh if you intend to keep reuse the JWKS over and over
		keySet, err := jwk.Fetch(context.Background(), formattedURLString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//parse token
		_, err = jwt.Parse(
			[]byte(splitAuthHeader[1]),
			jwt.WithKeySet(keySet),
			jwt.WithValidate(true),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}
