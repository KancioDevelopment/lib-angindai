package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/go-gomail/gomail"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	// MimeTypeHTML is ...
	MimeTypeHTML MimeType = "text/html"
)

var (
	mailer       *gomail.Dialer
	sendGridInit *sendgrid.Client
)

// MimeType is ...
type MimeType string

// MailOption is a ...
type MailOption struct {
	From         string
	To           string
	Cc           string
	Subject      string
	BodyMimeType MimeType
	Attachment   string
	Dialer       MailDialer
	SenderName   string
}

// MultiMailOption is a ...
type MultiMailOption struct {
	From
	Subject      string
	To           []*mail.Email
	Cc           []*mail.Email
	BodyMimeType MimeType
}

// From is ...
type From struct {
	Name  string
	Email string
}

// Tos is ...
type Tos struct {
	To []From
}

// MailDialer may be used if we use our own mail server.
// Currently we use sendgrid, so this is actually not needed.
// But we not ditch this code yet since we think we may build
// our own mail server using postfix later.
type MailDialer struct {
	Host     string
	Port     int
	Username string
	Password string
}

type NewDialerOptions struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Auth           smtp.Auth
	SSL            bool
	TLSConfig      *tls.Config
	LocalName      string
	SendGridApiKey string
	Sender         string
	SenderName     string
}

type Mail struct {
	From    string
	To      string
	Subject string
	Message string
}

func Init(dialOptions NewDialerOptions) {
	mailer = gomail.NewDialer(dialOptions.Host, dialOptions.Port, dialOptions.Username, dialOptions.Password)
	sendGridInit = sendgrid.NewSendClient(dialOptions.SendGridApiKey)
}

func SendByGoMail(mail Mail) error {
	m := gomail.NewMessage()
	m.SetHeader("From", mail.From)
	m.SetHeader("To", mail.To)
	m.SetHeader("Subject", mail.Subject)
	m.SetBody("text/html", mail.Message)
	if err := mailer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func (t *Tos) GenerateMail() []*mail.Email {
	var personalize []*mail.Email
	for _, data := range t.To {
		m := mail.NewEmail(data.Name, data.Email)
		personalize = append(personalize, m)
	}
	return personalize
}

func SendMail(message string, mailOpts MailOption) error {
	// We found a strange behavior of sendgrid. So there are 2 types
	// of content. Plain text and html. We must supply both value,
	// but it will send html content only. We will explore on this
	// case later. Since we need html content, then this is not an issue.
	var plainTextContent string = message
	var htmlContent string = message

	// Commented out this code because of strange behavior of
	// sendgrid. See above.

	// if mailOpts.BodyMimeType == MimeTypeHtml {
	// 	htmlContent = message
	// } else {
	// 	plainTextContent = message
	// }

	from := mail.NewEmail(mailOpts.SenderName, mailOpts.From)
	subject := mailOpts.Subject
	to := mail.NewEmail("", mailOpts.To)
	mailMessage := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	response, err := sendGridInit.Send(mailMessage)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("sendmail: Mail server response with status code: %d, response body: %s", response.StatusCode, response.Body)
	}

	return nil
}

func SendMultiMail(message string, mailOpts MultiMailOption) error {
	m := mail.NewV3Mail()
	from := mail.NewEmail(mailOpts.From.Name, mailOpts.From.Email)
	m.SetFrom(from)
	m.Subject = mailOpts.Subject

	p := mail.NewPersonalization()
	p.AddTos(mailOpts.To...)
	p.AddCCs(mailOpts.Cc...)
	p.Subject = mailOpts.Subject
	m.AddPersonalizations(p)

	var mimeType = fmt.Sprintf("%v", mailOpts.BodyMimeType)
	c := mail.NewContent(mimeType, message)
	m.AddContent(c)
	response, err := sendGridInit.Send(m)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("sendmail: Mail server response with status code :%d, response body: %s", response.StatusCode, response.Body)
	}
	return nil
}
