package services

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	sqlc "go.mod/internal/sqlc/generate"
	"go.mod/internal/utils"
	"golang.org/x/crypto/bcrypt"
)


type PublicService struct {
	queries *sqlc.Queries
	redis *redis.Client
}

func NewPublicService(queriespool *sqlc.Queries, redisclient *redis.Client) *PublicService {
	return &PublicService{queries: queriespool, redis: redisclient}
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
type ResetPass struct {
	Token string
	NewPass string
	ConfirmPass string
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

	err = s.SendConfirmEmail(ctx, signupData.Email)
	if err != nil {
		return err
	}

	return nil
}

func (s *PublicService) SendConfirmEmail(ctx *gin.Context, email string) (error) {

	// get user data from database
	userData, err := s.queries.GetUserData(ctx, email)
	if err != nil {
		return errors.New("not able to fetch user data from database")
	}

	// generate confirmation token
	confirmtokenData := utils.Token{
		Issuer: "signupFunc@PMS",	
		Subject: "confirm_token",
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		IssuedAt: time.Now().Unix(),
		Email: userData.Email,
	}
	confirm_token, err := utils.GenerateJWT(confirmtokenData)
	if err != nil {
		return errors.New("error generating confirm token. try again")
	}

	// generate confirmation link
	confirmationLink := fmt.Sprintf("%s/public/confirmsignup?token=%s", os.Getenv("Domain"), confirm_token)
	resendLink := fmt.Sprintf("%s/public/sendconfirmemail?email=%s", os.Getenv("Domain"), userData.Email) // TODO: this can be exploited 
																										// change the resend link logic

	// send confirmation email
	emailData := utils.EmailData{
		Email: userData.Email,
		Signup_Confirmation_Link: confirmationLink,
		Resend_Email_Link: resendLink,
		PathToTemplate: "./template/emails/confirmsignup.html",
		To_Email: []string{userData.Email},
	}
	go utils.SendEmailHTML(emailData)

	return nil
}

func (s *PublicService) ConfirmEmail(ctx *gin.Context, confirmToken string) (error) {
	
	// parse token
	claims, err := utils.ParseJWT(confirmToken)
	if err != nil {
		return errors.New("error parsing confirm token. please request a new link")
	}

	// get email from claims
	userEmail := claims["email"].(string)

	// update confirmed in the db
	err = s.queries.UpdateEmailConfirmation(ctx, userEmail)
	if err != nil {
		return errors.New("error updating email validity")
	}

	return nil
}

func (s *PublicService) SendResetPassEmail(ctx *gin.Context, email string) (error) {

	// get user data from database
	userData, err := s.queries.GetUserData(ctx, email)
	if err != nil {
		return errors.New("not able to fetch user data from database. enter valid email address")
	}

	// generate confirmation token
	resettokenData := utils.Token{
		Issuer: "resetpassFunc@PMS",	
		Subject: "reset_token",
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		IssuedAt: time.Now().Unix(),
		Email: userData.Email,
	}
	reset_token, err := utils.GenerateJWT(resettokenData)
	if err != nil {
		return errors.New("error generating confirm token. try again")
	}

	// generate confirmation link
	resetpassLink := fmt.Sprintf("%s/public/resetpassgetpass?token=%s", os.Getenv("Domain"), reset_token)
	
	// send confirmation email
	emailData := utils.EmailData{
		Email: userData.Email,
		Password_Reset_Link: resetpassLink,
		PathToTemplate: "./template/emails/resetpass.html",
		To_Email: []string{userData.Email},
	}
	go utils.SendEmailHTML(emailData)

	// add token as SetEx to redis db to avoid multiple requests 
	err = s.redis.SetEx(ctx, reset_token, userData.UserID, time.Minute * 15).Err()
	if err != nil {
		return errors.New("failed to add token to redis")
	}

	return nil
}

func (s *PublicService) GetPass(ctx *gin.Context) (bytes.Buffer, error) {

	var data ResetPass
	data.Token = ctx.Query("token")

	// check if link already used
	already_used, err := s.redis.Exists(ctx, data.Token).Result()
	if err != nil || already_used == 0 {
		return bytes.Buffer{}, errors.New("link already used. generate new link please")
	}

	body, err := utils.DynamicHTML("./template/reset/passresetpostpass.html", data)
	if err != nil {
		return bytes.Buffer{}, errors.New("failed to generate dynamic html")
	}

	return body, nil
}

func (s *PublicService) ResetPass(ctx *gin.Context, data ResetPass) (error) {

	// check if link already used
	already_used, err := s.redis.Exists(ctx, data.Token).Result()
	if err != nil || already_used == 0 {
		return errors.New("link already used. generate new link please")
	}

	// TODO: implement better input validation
	if data.NewPass != data.ConfirmPass {
		return errors.New("newpass and confirmpass do not match")
	}

	// parse token
	claims, err := utils.ParseJWT(data.Token)
	if err != nil {
		return errors.New("error parsing confirm token. please request a new link")
	}
	// get email from claims
	userEmail := claims["email"].(string)

	// hash the password
	hashed_pass, err := bcrypt.GenerateFromPassword([]byte(data.NewPass), 10)
	if err != nil {
		return errors.New("invalid password. try again")
	}	
	newPassString := string(hashed_pass)
	
	// remove token from redis db
	err = s.redis.Del(ctx, data.Token).Err()
	if err != nil {
		return errors.New("error deleting token. please request a new link")
	}

	// update password in the db
	err = s.queries.UpdatePassword(ctx, sqlc.UpdatePasswordParams{
		Password: newPassString,
		Email: userEmail,
	})
	if err != nil {
		return errors.New("error resetting password")
	}

	return nil
}

func (s *PublicService) LoginPost(ctx *gin.Context, loginData UserInputData) (int64, JWTTokens, error) {

	// check if user in database
	// if present, get all data from database
	userData, err := s.queries.GetUserData(ctx, loginData.Email)
	if err != nil {
		return 0, JWTTokens{}, errors.New("user does not exist. signup first")
	}

	if !userData.Confirmed {
		return 0, JWTTokens{}, errors.New("please verify email first")
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
