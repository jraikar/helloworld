package auth

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"net/http"
)

type CognitoRouter struct {
	Client CognitoClient
}

func NewCognitoRouter(client CognitoClient) *CognitoRouter {
	return &CognitoRouter{Client: client}
}

//SignupRequest
// swagger:model
type SignupRequest struct {
	// Username for signup
	// required : true
	// example: dev1
	Username string `json:"username"`
	// Password for the user
	// required : true
	// example: password1
	Password string `json:"password"`
}

//SignIn - signs up the user
// swagger:operation POST /api/v1/signup user-service user-service
// ---
//  summary: Signs up the user in aerostation
//  description: Signs up the user in aerostation
//  operationId: signup
//  produces:
//    - application/json
//  parameters:
//   - name: body
//     in: body
//     description: Request body for sign up
//     required: true
//     schema:
//	     $ref: '#/definitions/SignupRequest'
//  responses:
//    '200':
//      description: signup complete !
//    '400':
//      description: Invalid request
//    '500':
//      description: User already exist
func (c *CognitoRouter) SignUp(w http.ResponseWriter, r *http.Request) {

	//parse input request
	req := SignupRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//signup request for AWS Cognito
	awsReq := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(c.Client.AppClientId),
		Username: aws.String(req.Username),
		Password: aws.String(req.Password),
		//SecretHash: :todo: recommended to use this for best practice
	}

	//make api call with signup request to cognito
	_, err = c.Client.SignUp(r.Context(), awsReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//auto confirm auth, todo: let auth confirm account using email, Ref - https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-email-phone-verification.html
	//For email verification - todo: need to create new User pool in aws cognito with email as required attribute (meaning adding email in request) and handle all the flow.
	confirmInput := &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(c.Client.UserPoolId),
		Username:   aws.String(req.Username),
	}

	_, err = c.Client.AdminConfirmSignUp(r.Context(), confirmInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("signup complete !"))
}

//SignInRequest
// swagger:model
type SignInRequest struct {
	// Username for signin
	// required : true
	// example: dev1
	Username string `json:"username"`
	// Password for signin
	// required : true
	// example: password1
	Password string `json:"password"`
}

//SignInResponse
// swagger:model
type SignInResponse struct {
	// The access token. This has bare minimum details about auth and not attribute, if another micro service needs to call this service, then we can make use of access_token
	// This is used in authentication
	AccessToken *string `json:"access_token"`

	// The expiration period of the authentication result in seconds.
	ExpiresIn int32 `json:"expires_in"`

	// The ID token contains all attributes set in the auth pool, right now we have just username enabled in userpool, hence we can fetch only username from id_token
	// In future we might add more attributes and use this as identifier for PII (Personal Identifying  Identification)
	IdToken *string `json:"id_token"`

	// The refresh token. This is used to retrieve new access_token and id_token. By default, this expires in 30 days
	RefreshToken *string `json:"refresh_token"`

	// The token type.
	TokenType *string `json:"token_type"`
}

//SignIn - signs in the user
// swagger:operation POST /api/v1/signin user-service user-service
// ---
//  summary: Signs in the user in aerostation
//  description: Signs in the user in aerostation
//  operationId: signin
//  produces:
//    - application/json
//  parameters:
//   - name: body
//     in: body
//     description: Request body for sign in
//     required: true
//     schema:
//	     $ref: '#/definitions/SignInRequest'
//  responses:
//    '200':
//      description: success
//      schema:
//         $ref: '#/definitions/SignInResponse'
//    '400':
//      description: Invalid request
//    '500':
//      description: Internal server error
func (c *CognitoRouter) SignIn(w http.ResponseWriter, r *http.Request) {

	//parse input request
	req := SignInRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	signInInput := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: "ADMIN_USER_PASSWORD_AUTH",
		AuthParameters: map[string]string{
			"USERNAME": req.Username,
			"PASSWORD": req.Password,
		},
		ClientId:   aws.String(c.Client.AppClientId),
		UserPoolId: aws.String(c.Client.UserPoolId),
	}

	output, err := c.Client.AdminInitiateAuth(r.Context(), signInInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := SignInResponse{
		AccessToken:  output.AuthenticationResult.AccessToken,
		ExpiresIn:    output.AuthenticationResult.ExpiresIn,
		IdToken:      output.AuthenticationResult.IdToken,
		RefreshToken: output.AuthenticationResult.RefreshToken,
		TokenType:    output.AuthenticationResult.TokenType,
	}

	_ = json.NewEncoder(w).Encode(res)
}
