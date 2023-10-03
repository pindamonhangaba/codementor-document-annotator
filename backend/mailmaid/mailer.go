package mailmaid

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/pindamonhangaba/monoboi/backend/service"
	"github.com/pkg/errors"
)

type mailTemplate[T any] string

func (f mailTemplate[T]) Parse(data T) (bytes.Buffer, error) {
	var body bytes.Buffer

	html, err := os.ReadFile(string(f))
	if err != nil {
		return body, errors.Wrap(err, "read file")
	}

	t, err := template.New(string(f)).Parse(string(html))
	if err != nil {
		return body, errors.Wrap(err, "parse hmtl")
	}
	if err := t.Execute(&body, data); err != nil {
		return body, errors.Wrap(err, "build html")
	}
	return body, nil
}

type availableTemplates struct {
	ResetPasswordHtml         mailTemplate[service.TwoFactorRequest]
	ResetPasswordLinkHtml     mailTemplate[service.TwoFactorRequest]
	ConfirmRegistrationHtml   mailTemplate[service.TwoFactorRequest]
	EmailAuthenticateHTML     mailTemplate[service.TwoFactorRequest]
	EmailAuthenticateLinkHTML mailTemplate[service.TwoFactorRequest]
	AlertNewDeviceLoginHTML   mailTemplate[service.AlertNewDeviceLogin]
}

type Mailer struct {
	availableTemplates
	Config MailerConf
}

func p[T ~string](t *T, s, f string) {
	n := s + string(os.PathSeparator) + string(f)
	*t = T(n)
}

func NewMailer(conf MailerConf) *Mailer {
	tmplts := availableTemplates{}
	p(&tmplts.ResetPasswordHtml, conf.TemplatePath, "reset-password.html")
	p(&tmplts.ResetPasswordLinkHtml, conf.TemplatePath, "reset-password-link.html")
	p(&tmplts.ConfirmRegistrationHtml, conf.TemplatePath, "confirm-registration.html")
	p(&tmplts.EmailAuthenticateHTML, conf.TemplatePath, "email-authenticate.html")
	p(&tmplts.EmailAuthenticateLinkHTML, conf.TemplatePath, "email-authenticate-link.html")
	p(&tmplts.AlertNewDeviceLoginHTML, conf.TemplatePath, "alert-new-device-login.html")

	return &Mailer{
		Config:             conf,
		availableTemplates: tmplts,
	}
}

type MailerConf struct {
	User,
	Password,
	From,
	TemplatePath string
}

// SendPwdResetRequest sending email to reset password
func (m *Mailer) SendPwdResetRequest(pr service.TwoFactorRequest, model service.AccessModel) error {

	parseTmplt := m.availableTemplates.ResetPasswordHtml.Parse
	if model == service.LinkAccessModel {
		parseTmplt = m.availableTemplates.ResetPasswordLinkHtml.Parse
	}

	body, err := parseTmplt(pr)
	if err != nil {
		return err
	}

	emailBody := body.String()
	subject := "Solicitação de recuperação de senha"

	to := pr.Name + " <" + pr.Email + ">"
	_, err = sendMail(m.Config.User, m.Config.Password, message{
		To:       to,
		From:     m.Config.From,
		Subject:  subject,
		Body:     emailBody,
		TextBody: "Código para recuperar sua senha: " + pr.PasswordNumber + "\r\n\r\nOu acesse: " + pr.ConfirmationURL,
	})
	return err
}

// SendAccountEmailConfirm sending email to confirm account
func (m *Mailer) SendAccountEmailConfirm(pr service.TwoFactorRequest) error {
	body, err := m.availableTemplates.ConfirmRegistrationHtml.Parse(pr)
	if err != nil {
		return err
	}

	subject := "Confirmação de email"
	emailBody := body.String()
	to := pr.Name + " <" + pr.Email + ">"
	_, err = sendMail(m.Config.User, m.Config.Password, message{
		To:       to,
		From:     m.Config.From,
		Subject:  subject,
		Body:     emailBody,
		TextBody: "Confirme seu email",
	})
	return err
}

func (m *Mailer) SendPwdResetAlert(config ...interface{}) error {
	return nil
}

// SendAuthenticationRequest sends an email to authenticate  a user
func (m *Mailer) SendAuthenticationRequest(pr service.TwoFactorRequest, model service.AccessModel) error {
	parseTmplt := m.availableTemplates.EmailAuthenticateHTML.Parse
	if model == service.LinkAccessModel {
		parseTmplt = m.availableTemplates.EmailAuthenticateLinkHTML.Parse
	}

	body, err := parseTmplt(pr)
	if err != nil {
		return err
	}

	emailBody := body.String()
	subject := "Autenticação de usuário"

	to := pr.Name + " <" + pr.Email + ">"
	_, err = sendMail(m.Config.User, m.Config.Password, message{
		To:       to,
		From:     m.Config.From,
		Subject:  subject,
		Body:     emailBody,
		TextBody: "Código para autenticação: " + pr.PasswordNumber + "\r\n\r\nOu acesse: " + pr.ConfirmationURL,
	})
	return err
}

type message struct {
	To          string
	From        string
	Subject     string
	Body        string
	TextBody    string
	Attachments []messageAttachment
}

type messageAttachment struct {
	URL  string
	Name string
}

func sendMail(user, pwd string, msg message) (string, error) {
	mg := mailgun.NewMailgun(user, pwd)
	m := mg.NewMessage(
		msg.From,
		msg.Subject,
		msg.TextBody,
		msg.To,
	)
	m.SetHtml(msg.Body)

	for _, f := range msg.Attachments {
		dat, err := os.Open(f.URL)
		if err != nil {
			return "failed attachments", errors.Wrap(err, "attachment "+f.URL)
		}
		m.AddReaderAttachment(f.Name, dat)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	msgRet, id, err := mg.Send(ctx, m)
	if err != nil {
		log.Println(msgRet, msg.To, msg.From, msg.Subject, msg.TextBody, id, err)
	}
	return id + "." + msgRet, err
}
