package utils

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
)

// TODO: this is stupid
type EmailData struct {
	Name string
	Email string
	Signup_Confirmation_Link string
	Resend_Email_Link string
	Password_Reset_Link string
}

func SendEmailHTML (body bytes.Buffer, 	to_Email []string) {
	
	// required headers for html
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	// construct the email string
	demo := "Subject: PMS" + "\n" + headers + "\n\n" + body.String()


	// initialise the Plain Authentication mechanism
	auth := smtp.PlainAuth(
		"",
		os.Getenv("SMTP_GO_Username"),
		os.Getenv("SMTP_GO_Pass"),
		os.Getenv("SMTP_GO_Host"),
	)
	// connect to server and send email
	err := smtp.SendMail(
		os.Getenv("SMTP_GO_HostAddress"),
		auth,
		os.Getenv("SMTP_GO_From"),
		to_Email,
		[]byte(demo),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

}