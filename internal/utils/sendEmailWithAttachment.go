package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"os"
	"path/filepath"
)


func SendEmailHTMLWithAttachment(body bytes.Buffer, to_Email []string, attachment *[]byte, name string) (error)  {
	// Email headers
	subject := "Subject: PMS\n"
	fromEmail := os.Getenv("SMTP_GO_From")

	// Load environment variables
	smtpHost := os.Getenv("SMTP_GO_Host")
	smtpPort := os.Getenv("SMTP_GO_HostAddress")
	username := os.Getenv("SMTP_GO_Username")
	password := os.Getenv("SMTP_GO_Pass")

	// Validate required environment variables
	if smtpHost == "" || smtpPort == "" || fromEmail == "" || username == "" || password == "" {
		fmt.Println("missing required SMTP configuration in environment variables")
	}

	// Construct the email
	var emailContent bytes.Buffer
	writer := multipart.NewWriter(&emailContent)

	// MIME headers for the email
	emailContent.WriteString(fmt.Sprintf("From: %s\n", fromEmail))
	emailContent.WriteString(fmt.Sprintf("To: %s\n", to_Email[0])) // Assuming only one recipient
	emailContent.WriteString(subject)
	emailContent.WriteString("MIME-Version: 1.0\n")
	emailContent.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n\n", writer.Boundary()))

	// Add the HTML body
	htmlPart, err := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/html; charset=UTF-8"},
	})
	if err != nil {
		fmt.Println("failed to create email body part: %w", err)
	}
	_, err = htmlPart.Write(body.Bytes())
	if err != nil {
		fmt.Println("failed to write email body: %w", err)
	}

	// Add the file attachment
	if attachment != nil {

		// Encode file in Base64
		encoded := base64.StdEncoding.EncodeToString(*attachment)

		// Create attachment part
		attachmentPart, err := writer.CreatePart(map[string][]string{
			"Content-Type":              {fmt.Sprintf("%s; name=%s", "application/octet-stream", name)},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=%s", name)},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			fmt.Println("failed to create attachment part: %w", err)
		}

		// Write encoded file content to the attachment part
		_, err = attachmentPart.Write([]byte(encoded))
		if err != nil {
			fmt.Println("failed to write attachment: %w", err)
		}
	}

	// Close the writer
	writer.Close()

	// Plain authentication
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Send the email
	err = smtp.SendMail(smtpPort, auth, fromEmail, to_Email, emailContent.Bytes())
	if err != nil {
		fmt.Println("failed to send email: %w", err)
	}
	return nil
}

func SendEmailHTMLWithAttachmentFileHeader(body bytes.Buffer, to_Email []string, fileHeader *multipart.FileHeader) (error) {

	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file content
	fileContent := new(bytes.Buffer)
	_, err = fileContent.ReadFrom(file)
	if err != nil {
		fmt.Println("failed to read file content: %w", err)
	}

	fileName := filepath.Base(fileHeader.Filename)
	fileBytes := fileContent.Bytes()

	return SendEmailHTMLWithAttachment(body, to_Email, &fileBytes, fileName)
}

func SendEmailHTMLWithAttachmentFileBytes(body bytes.Buffer, to_Email []string, fileBytes []byte, fileName string) (error) {
	return SendEmailHTMLWithAttachment(body, to_Email, &fileBytes, fileName)
}
func SendEmailHTMLWithAttachmentFilePath(body bytes.Buffer, to_Email []string, filePath string, fileName string) (error) {

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return SendEmailHTMLWithAttachment(body, to_Email, &fileBytes, fileName)
}