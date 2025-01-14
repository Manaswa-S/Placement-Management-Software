package services

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.mod/internal/config"
	"go.mod/internal/dto"
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
	confirmtokenData := dto.Token{
		Issuer: "signupFunc@PMS",	
		Subject: "confirm_token",
		ExpiresAt: time.Now().Add(config.SignupConfirmLinkTokenExpiration * time.Minute).Unix(),
		IssuedAt: time.Now().Unix(),
		Email: userData.Email,
		Role: userData.Role,
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
	}
	template, err := utils.DynamicHTML("./template/emails/confirmsignup.html", emailData)
	if err != nil {
		return err
	}
	go utils.SendEmailHTML(template, []string{userData.Email})

	return nil
}

func (s *PublicService) ConfirmEmail(ctx *gin.Context, confirmToken string) (bytes.Buffer, error) {
	
	// parse token
	claims, err := utils.ParseJWT(confirmToken)
	if err != nil {
		return bytes.Buffer{}, errors.New("error parsing confirm token. please request a new link")
	}

	// get email from claims
	userEmail := claims["email"].(string)

	// update confirmed in the db
	err = s.queries.UpdateEmailConfirmation(ctx, userEmail)
	if err != nil {
		return bytes.Buffer{}, errors.New("error updating email validity")
	}

	// embed token in email
	pathtoHTML := "./template/extrainfoforms/companyform.html"
	if int64(claims["role"].(float64)) == 1 {
		pathtoHTML = "./template/extrainfoforms/studentform.html"
	}

	body, err := utils.DynamicHTML(pathtoHTML, ResetPass{Token: confirmToken})
	if err != nil {
		return bytes.Buffer{}, errors.New("failed to generate dynamic html")
	}

	return body, nil
}

func (s *PublicService) SendResetPassEmail(ctx *gin.Context, email string) (error) {

	// get user data from database
	userData, err := s.queries.GetUserData(ctx, email)
	if err != nil {
		return err
	}

	// generate confirmation token
	resettokenData := dto.Token{
		Issuer: "resetpassFunc@PMS",	
		Subject: "reset_token",
		ExpiresAt: time.Now().Add(config.ResetLinkTokenExpiration * time.Minute).Unix(),
		IssuedAt: time.Now().Unix(),
		Email: userData.Email,
	}

	reset_token, err := utils.GenerateJWT(resettokenData)
	if err != nil {
		return err
	}

	// generate confirmation link
	resetpassLink := fmt.Sprintf("%s/public/resetpassgetpass?token=%s", os.Getenv("Domain"), reset_token)
	
	// send confirmation email
	emailData := utils.EmailData{
		Email: userData.Email,
		Password_Reset_Link: resetpassLink,
	}
	template, err := utils.DynamicHTML("./template/emails/resetpass.html", emailData)
	if err != nil {
		return err
	}
	go utils.SendEmailHTML(template, []string{userData.Email})

	// add token as SetEx to redis db to avoid multiple requests 
	err = s.redis.SetEx(ctx, reset_token, userData.UserID, time.Minute * 15).Err()
	if err != nil {
		return err
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

func (s *PublicService) LoginPost(ctx *gin.Context, loginData UserInputData) (int64, *dto.JWTTokens, error) {
	tokens := &dto.JWTTokens{}
	// check if user in database
	// if present, get all data from database
	userData, err := s.queries.GetUserData(ctx, loginData.Email)
	if err != nil {
		return 0, nil, errors.New("user does not exist. signup first")
	}

	if !userData.Confirmed {
		return 0, nil, errors.New("please verify email first")
	}

	if !userData.IsVerified {
		return 0, nil, errors.New("user verification is pending")
	}

	// compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(loginData.Password))
	if err != nil {
		return 0, nil, errors.New("password does not match. try again")
	}


	// generate JWT access token and refresh token
	accesstokenData := dto.Token{
			Issuer: "loginFunc@PMS",	
			Subject: "access_token",
			ExpiresAt: time.Now().Add(config.JWTAccessExpiration * time.Minute).Unix(),
			IssuedAt: time.Now().Unix(),
			Role: userData.Role,
			ID: userData.UserID,
	}
	access_token, err := utils.GenerateJWT(accesstokenData)
	if err != nil {
		return 0, nil, errors.New("error generating access token. try again")
	}

	// generate JWT access token and refresh token
	refreshtokenData := dto.Token{
		Issuer: "loginFunc@PMS",	
		Subject: "refresh_token",
		ExpiresAt: time.Now().Add(config.JWTRefreshExpiration * time.Hour).Unix(),
		IssuedAt: time.Now().Unix(),
		Role: userData.Role,
		ID: userData.UserID,
	}
	refresh_token, err := utils.GenerateJWT(refreshtokenData)
	if err != nil {
		return 0, nil, errors.New("error generating refresh token. try again")
	}

	tokens.JWTAccess = access_token
	tokens.JWTRefresh = refresh_token

	// return the jwt tokens and any errors
	return userData.Role, tokens, nil
}

func (s *PublicService) ExtraInfoPostStudent(ctx *gin.Context, claims jwt.MapClaims) (sqlc.Student, error) {
	// bind data
	var data dto.ExtraInfoStudent
	err := ctx.Bind(&data)
	if err != nil {
		return sqlc.Student{}, err
	}
	// get resume file
	resumeFile, err := ctx.FormFile("Resume")
	if err != nil {
		return sqlc.Student{}, err
	}
	// save resume file
	resumeStoragePath := fmt.Sprintf("%s%s&%d%s", os.Getenv("ResumeStorageDir"), data.CollegeRollNumber, time.Now().Unix(), filepath.Ext(resumeFile.Filename))
	resumeSavePath, err := utils.SaveFile(ctx, resumeStoragePath, resumeFile)
	if err != nil {
		return sqlc.Student{}, err
	}
	// get result file
	resultFile, err := ctx.FormFile("Result")
	if err != nil {
		return sqlc.Student{}, errors.New("unable to get result file. try again")
	}
	// save result file
	resultStoragePath := fmt.Sprintf("%s%s&%d%s", os.Getenv("ResultStorageDir"), data.CollegeRollNumber, time.Now().Unix(), filepath.Ext(resultFile.Filename))
	resultSavePath, err := utils.SaveFile(ctx, resultStoragePath, resultFile)
	if err != nil {
		return sqlc.Student{}, errors.New("unable to save result file. try again")
	}
	// update data in db
	data.StudentEmail = claims["email"].(string)
	studentData, err := s.queries.ExtraInfoStudent(ctx, sqlc.ExtraInfoStudentParams{
		StudentName: data.StudentName,
		RollNumber: data.CollegeRollNumber,
		StudentDob: pgtype.Date{Time: data.DateOfBirth, Valid: true},
		Gender: data.Gender,
		Course: data.Course,
		Department: data.Department,
		YearOfStudy: data.YearOfStudy,
		ResumeUrl: pgtype.Text{String: resumeSavePath, Valid: true},
		ResultUrl: resultSavePath,
		Cgpa: pgtype.Float8{Float64: data.CGPA, Valid: true},
		ContactNo: data.ContactNumber,
		StudentEmail: data.StudentEmail,
		Address: pgtype.Text{String: data.Address, Valid: true},
		Skills: pgtype.Text{String: data.Skills, Valid: true},
		Email: data.StudentEmail,
	})
	if err != nil {
		return sqlc.Student{}, err
	}

	// return
	return studentData, nil
}

func (s *PublicService) ExtraInfoPostCompany(ctx *gin.Context, claims jwt.MapClaims) (sqlc.Company, error) {
	// bind incoming data
	var data dto.ExtraInfoCompany
	err := ctx.Bind(&data)
	if err != nil {
		return sqlc.Company{}, errors.New("unable to bind company data. try again")
	}
	// update db
	data.RepresentativeEmail = claims["email"].(string)
	companyData, err := s.queries.ExtraInfoCompany(ctx, sqlc.ExtraInfoCompanyParams{
		CompanyName: data.CompanyName,
		RepresentativeEmail: data.RepresentativeEmail,
		RepresentativeContact: data.RepresentativeContact,
		RepresentativeName: data.RepresentativeName,
		DataUrl: pgtype.Text{String: ""},
		Email: data.RepresentativeEmail,
	})
	if err != nil {
		return sqlc.Company{}, errors.New("unable to update company data in database")
	}
	// return
	return companyData, nil
}



