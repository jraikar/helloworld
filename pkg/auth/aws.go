package auth

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"log"
	"os"
)

//CognitoClient - holds aws cognito client with app client id and auth pool id
type CognitoClient struct {
	AppClientId string
	UserPoolId  string
	*cognitoidentityprovider.Client
}

// Init - This loads the aws config passed in and returns aws cognito client
func Init() (*CognitoClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Println("unable to load aws client", err)
		return nil, err
	}
	return &CognitoClient{
		AppClientId: os.Getenv("COGNITO_APP_CLIENT_ID"),
		UserPoolId:  os.Getenv("COGNITO_USER_POOL_ID"),
		Client:      cognitoidentityprovider.NewFromConfig(cfg),
	}, nil
}
