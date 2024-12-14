package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
)


type EmailData struct {
	Name string
	Email string
	Signup_Confirmation_Link string
	Resend_Email_Link string
	Password_Reset_Link string
	// Below 2 fields are always required
	PathToTemplate string
	To_Email []string
}

func SendEmailHTML (data EmailData) {

	var body bytes.Buffer

	// Parse the template file into object assigned to 'bodytemp'
	bodytemplate, err := template.ParseFiles(data.PathToTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Execute the template and apply 'data' to the template
	// store the formed result in 'body'
	err = bodytemplate.Execute(&body, data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	
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
	err = smtp.SendMail(
		os.Getenv("SMTP_GO_HostAddress"),
		auth,
		os.Getenv("SMTP_GO_From"),
		data.To_Email,
		[]byte(demo),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

}