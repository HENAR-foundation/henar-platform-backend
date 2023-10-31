package email

import (
	"fmt"
	"log"
	"os"

	"henar-backend/sentry"
	"henar-backend/types"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
	"github.com/mailjet/mailjet-apiv3-go/v4/resources"
)

type MailjetClient struct {
	client *mailjet.Client
	sender []resources.Sender
}

func Init() *MailjetClient {
	publicKey := os.Getenv("MAILJET_APIKEY_PUBLIC")
	secretKey := os.Getenv("MAILJET_APIKEY_PRIVATE")

	m := mailjet.NewMailjetClient(publicKey, secretKey)

	var data []resources.Sender
	count, _, err := m.List("sender", &data)
	if err != nil {
		log.Printf("Unexpected error during email sender initiation: %s", err)
	}
	if count < 1 {
		log.Printf("At least one sender expected !")
	}

	return &MailjetClient{
		client: m,
		sender: data,
	}
}

func (c *MailjetClient) SendEmail(toEmail, name, subject, textPart, htmlPart string) error {

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: c.sender[0].Email,
				Name:  c.sender[0].Name,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: toEmail,
					Name:  name,
				},
			},
			Subject:  subject,
			TextPart: textPart,
			HTMLPart: htmlPart,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := c.client.SendMailV31(&messages)
	if err != nil {
		sentry.SentryHandler(err)
	}

	return err
}

func (c *MailjetClient) SendConfirmationEmail(verificationData types.VerificationData) error {
	subject := "Confirmation Email"
	verifyUrl := fmt.Sprintf("https://healthnet.am/verify-email/%s", verificationData.Code)
	textPart := fmt.Sprintf("Hello! Thank you for joining Henar! Click the following link to confirm your email:  %s", verifyUrl)
	htmlPart := fmt.Sprintf(`<p>Hello! Thank you for joining Henar!</p><p>Click the following link to confirm your email: <a href="%s">%s</a></p><p>If you didn't requested this email please ignore it.</p> <br><br><p>Henar Foundation</p>`, verifyUrl, verifyUrl)

	return c.SendEmail(verificationData.Email, "Recipient", subject, textPart, htmlPart)
}
