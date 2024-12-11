package services

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
	"golang.org/x/crypto/bcrypt"
)


type PublicService struct {
	queries *sqlc.Queries
}

func NewPublicService(queriespool *sqlc.Queries) *PublicService {
	return &PublicService{queries: queriespool}
}


// defined structs
type UserInputData struct {
	Email string
	Password string
}
type JWTTokens struct {
	JWTAccess string
	JWTRefresh string
}

func (s *PublicService) SignupPost(ctx *gin.Context, signupData sqlc.SignupUserParams) (error) {

	// check if both email and password are valid
	// implement better validation function later on
	if signupData.Email == "" || signupData.Password == "" {
		return errors.New("invalid email or password. try again")
	}

	// hash the password
	hashed_pass, err := bcrypt.GenerateFromPassword([]byte(signupData.Password), 10)
	if err != nil {
		return errors.New("invalid password. try again")
	}
	signupData.Password = string(hashed_pass)

	// check if user in DB
	// if not register new user
	_, err = s.queries.SignupUser(ctx, signupData)
	if err != nil {
		return errors.New("user with email id already exists. try with different id")
	}

	// need to send the dashboard directly without the user needing to login again
	return nil
}


func (s *PublicService) LoginPost(ctx *gin.Context, loginData UserInputData) (int64, JWTTokens, error) {

	// check if user in database
	// if present, get all data from database
	userData, err := s.queries.LoginUser(ctx, loginData.Email)
	if err != nil {
		return 0, JWTTokens{}, errors.New("user does not exist. signup first")
	}

	// compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(loginData.Password))
	if err != nil {
		return 0, JWTTokens{}, errors.New("password does not match. try again")
	}


	// generate JWT access token and refresh token
	accesstokenData := utils.Token{
			Issuer: "loginFunc@PMS",	
			Subject: "access_token",
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
			IssuedAt: time.Now().Unix(),
			Role: userData.Role,
			ID: userData.UserID,
	}
	access_token, err := utils.GenerateJWT(accesstokenData)
	if err != nil {
		return 0, JWTTokens{}, errors.New("error generating access token. try again")
	}

	// generate JWT access token and refresh token
	refreshtokenData := utils.Token{
		Issuer: "loginFunc@PMS",	
		Subject: "refresh_token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
		IssuedAt: time.Now().Unix(),
		Role: userData.Role,
		ID: userData.UserID,
	}
	refresh_token, err := utils.GenerateJWT(refreshtokenData)
	if err != nil {
		return 0, JWTTokens{}, errors.New("error generating refresh token. try again")
	}


	// return the jwt tokens and any errors
	return userData.Role, JWTTokens{JWTAccess: access_token, JWTRefresh: refresh_token}, nil
}