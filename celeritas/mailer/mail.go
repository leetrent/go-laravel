package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"time"

	apimail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

// Mail holds the information necessary to connect to an SMTP server
type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIUrl      string
}

// Message is the type for an email message
type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

// Result contains information regarding the status of the sent email message
type Result struct {
	Success bool
	Error   error
}

// ListenForMail listens to the mail channel and sends mail
// when it receives a payload. It runs continually in the background,
// and sends error/success messages back on the Results channel.
// Note that if api and api key are set, it will prefer using
// an api to send mail
func (m *Mail) ListenForMail() {
	snippet := "[celeritas][mail.go][ListenForMail] =>"
	for {
		msg := <-m.Jobs
		fmt.Println("")
		fmt.Printf("%s (msg): %s", snippet, msg)
		fmt.Println("")
		err := m.Send(msg)
		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

func (m *Mail) Send(msg Message) error {
	fmt.Println("")
	fmt.Println("[celeritas][mail.go][Send] =>")
	fmt.Println("")

	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		m.ChooseAPI(msg)
	}
	return m.SendSMTPMessage(msg)
}

func (m *Mail) ChooseAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("unkown api: '%s'; only mailgun, sparkpost and sendgrid are supported", m.API)
	}
}

func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(transport, cfg)
	if err != nil {
		return err
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachements []apimail.Attachment

		for _, x := range msg.Attachments {
			var attach apimail.Attachment
			content, err := ioutil.ReadFile(x)
			if err != nil {
				return err
			}
			fileName := filepath.Base(x)
			attach.Bytes = content
			attach.Filename = fileName
			attachements = append(attachements, attach)
		}
		tx.Attachments = attachements
	}
	return nil
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	fmt.Println("")
	fmt.Println("[celeritas][mail.go][SendSMTPMessage] =>")
	fmt.Println("")

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	logSnippet := "\n[celeritas][mail.go][SendSMTPMessage] =>"
	fmt.Printf("%s (m.Host)...............: %s", logSnippet, m.Host)
	fmt.Printf("%s (server.Host)..........: %s", logSnippet, server.Host)
	fmt.Printf("%s (server.Port)..........: %d", logSnippet, server.Port)
	fmt.Printf("%s (server.Username)......: %s", logSnippet, server.Username)
	fmt.Printf("%s (server.Password)......: %s", logSnippet, server.Password)
	fmt.Printf("%s (server.KeepAlive).....: %t", logSnippet, server.KeepAlive)
	fmt.Printf("%s (server.ConnectTimeout): %v", logSnippet, server.ConnectTimeout)
	fmt.Printf("%s (server.SendTimeout)...: %v", logSnippet, server.SendTimeout)

	fmt.Println("\n[celeritas][mail.go][SendSMTPMessage] => (Calling server.Connect...)")

	smptClient, err := server.Connect()
	if err != nil {
		fmt.Println("")
		fmt.Printf("[celeritas][mail.go][SendSMTPMessage] => (server.Connect error): %s", err)
		fmt.Println("")
		return err
	}

	fmt.Println("[celeritas][mail.go][SendSMTPMessage] => (Returning from server.Connect...)")

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	fmt.Println("[celeritas][mail.go][SendSMTPMessage] => (Calling email.Send...)")
	err = email.Send(smptClient)
	if err != nil {
		fmt.Println("")
		fmt.Printf("[celeritas][mail.go][SendSMTPMessage] => (email.Send error): %s", err)
		fmt.Println("")
		return err
	}

	fmt.Println("[celeritas][mail.go][SendSMTPMessage] => (Returning from email.Send...)")

	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inLineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

func (m *Mail) inLineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
