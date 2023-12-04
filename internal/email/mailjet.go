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
	host   string
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
		host:   "https://healthnet.am",
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
	verifyUrl := fmt.Sprintf("%s/verify-email/%s", c.host, verificationData.Code)
	textPart := fmt.Sprintf("Hello! Thank you for joining Henar! Click the following link to confirm your email:  %s", verifyUrl)
	htmlPart := fmt.Sprintf(`
	<p>Hello!</p>
	<p>Welcome to HealthNet! We’re delighted to have you join our community of medical professionals in Armenia.</p>
	<p>To begin, please verify your email address by clicking on the link below. This step will grant you full access to our platform:</p>
	<p><a href="%s">[Verification Link]</a></p>
	<p>For any questions, reach out to our support team at <a href="mailto: info@healthnet.am">info@healthnet.am</a></p>
	<p>Glad to have you with us!</p>
	<p>Best regards,</p>
	<p>The HealthNet Team</p>
	<br>
	<p>Բարև ձեզ</p>
	<p>Բարի գալուստ HealthNet! Ուրախ ենք, Հայաստանի առողջապահության ոլորտի մասնագետների մեր համայնքին միանալու համար:</p>
	<p>Սկսելու համար խնդրում ենք հաստատել ձեր էլ․ հասցեն՝ սեղմելով ստորև նշված հղումը: Այս քայլը ձեզ հնարավորություն կտա լիարժեք մուտք գործել մեր հարթակ.</p>
	<p><a href="%s">[Verification Link]</a></p>
	<p>Հարցերի դեպքում դիմեք մեր աջակցման թիմին <a href="mailto: info@healthnet.am">info@healthnet.am</a>:</p>
	<p>Ուրախ ենք, որ մեզ հետ եք:</p>
	<p>Հարգանքով,</p>
	<p>HealthNet թիմ</p>`,
		verifyUrl, verifyUrl)

	return c.SendEmail(verificationData.Email, "Recipient", subject, textPart, htmlPart)
}

func (c *MailjetClient) SendPasswordResetEmail(verificationData types.VerificationData) error {
	subject := "Password Reset Request for Henar"
	resetUrl := fmt.Sprintf("%s/reset-password/%s", c.host, verificationData.Code)
	textPart := "Hello! We received a request to reset the password for your account. If you made this request, please click the link below to reset your password:"
	htmlPart := fmt.Sprintf(`
	<p>Hello!</p>
	<p>To ensure the security of your HealthNet account, you’ve requested a password reset. Follow the instructions below to create a new password:</p>
	<p>Click on the link below to reset your password:</p>
	<p><a href="%s">[Password Reset Link]</a></p>
	<p>Enter your new password. Make sure it's secure and unique.</p>
	<p>Your account security matters to us. If you didn't request this change, please contact our support team immediately at <a href="mailto: info@healthnet.am">info@healthnet.am</a></p>
	<p>Thank you for choosing HealthNet.</p>
	<p>Best regards,</p>
	<p>The HealthNet Team</p>
	<br>
	<p>Բարև ձեզ</p>
	<p>Ձեր HealthNet հաշվի անվտանգությունն ապահովելու համար Դուք դիմել եք գաղտնաբառի վերականգման համար: Նոր գաղտնաբառ ստեղծելու համար հետևեք ստորև ներկայացված հրահանգներին.</p>
	<p>Ձեր գաղտնաբառը վերականգնելու համար սեղմեք ստորև նշված հղումը.</p>
	<p><a href="%s">[Password Reset Link]</a></p>
	<p>Մուտքագրեք նոր գաղտնաբառը: Համոզվեք, որ այն անվտանգ է:</p>
	<p>Մեզ համար կարևոր է ձեր հաշվի անվտանգությունը: Եթե դուք չեք դիմել այս փոփոխության համար, խնդրում ենք անմիջապես կապվել մեր աջակցման թիմին <a href="mailto: info@healthnet.am">info@healthnet.am</a> հասցեով:</p>
	<p>Շնորհակալություն HealthNet-ն ընտրելու համար:</p>
	<p>Հարգանքով,</p>
	<p>HealthNet թիմ</p>
	`,
		resetUrl, resetUrl)

	return c.SendEmail(verificationData.Email, "Recipient", subject, textPart, htmlPart)
}
