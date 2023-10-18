package email

import (
	"fmt"
	"log"
	"os"

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

	mj := mailjet.NewMailjetClient(publicKey, secretKey)
	var sender []resources.Sender

	return &MailjetClient{
		client: mj,
		sender: sender,
	}
}

func (c *MailjetClient) SendEmail(toEmail, subject, textPart, htmlPart string) error {

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: c.sender[0].Email,
				Name:  c.sender[0].Name,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: toEmail,
					Name:  "",
				},
			},
			Subject:  subject,
			TextPart: textPart,
			HTMLPart: htmlPart,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	res, err := c.client.SendMailV31(&messages)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Email Data: %+v\n", res)
	return err
}
