// Package txemail sends transactional emails.
package txemail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strconv"

	"github.com/jordan-wright/email"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Message describes an email message to be sent, aliased in this package for convenience.
type Message = txtypes.Message

var emailSendCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_email_send",
	Help: "Number of emails sent.",
}, []string{"success"})

// render returns the rendered message contents without sending email.
func render(message Message) (*email.Email, error) {
	m := email.Email{
		To:      message.To,
		From:    conf.Get().EmailAddress,
		Headers: make(textproto.MIMEHeader),
	}
	if message.ReplyTo != nil {
		m.ReplyTo = []string{*message.ReplyTo}
	}
	if message.MessageID != nil {
		m.Headers["Message-ID"] = []string{*message.MessageID}
	}
	if message.FromName != "" {
		m.From = message.FromName
	}
	if len(message.References) > 0 {
		// jordan-wright/email does not support lists, so we must build it ourself.
		var refsList string
		for _, ref := range message.References {
			if refsList != "" {
				refsList += " "
			}
			refsList += fmt.Sprintf("<%s>", ref)
		}
		m.Headers["References"] = []string{refsList}
	}

	parsed, err := ParseTemplate(message.Template)
	if err != nil {
		return nil, err
	}

	if err := renderTemplate(parsed, message.Data, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// Send sends a transactional email.
//
// Callers that do not live in the frontend should call internalapi.Client.SendEmail
// instead. TODO(slimsag): needs cleanup as part of upcoming configuration refactor.
func Send(ctx context.Context, message Message) (err error) {
	if MockSend != nil {
		return MockSend(ctx, message)
	}
	if disableSilently {
		return nil
	}

	defer func() {
		emailSendCounter.WithLabelValues(strconv.FormatBool(err == nil)).Inc()
	}()

	conf := conf.Get()
	if conf.EmailAddress == "" {
		return errors.New("no \"From\" email address configured (in email.address)")
	}
	if conf.EmailSmtp == nil {
		return errors.New("no SMTP server configured (in email.smtp)")
	}

	m, err := render(message)
	if err != nil {
		return errors.Wrap(err, "render")
	}
	raw, err := m.Bytes()
	if err != nil {
		return errors.Wrap(err, "get bytes")
	}

	// Disable Mandrill features, because they make the emails look sketchy.
	if conf.EmailSmtp.Host == "smtp.mandrillapp.com" {
		// Disable click tracking ("noclicks" could be any string; the docs say that anything will disable click tracking except
		// those defined at
		// https://mandrill.zendesk.com/hc/en-us/articles/205582117-How-to-Use-SMTP-Headers-to-Customize-Your-Messages#enable-open-and-click-tracking).
		m.Headers["X-MC-Track"] = []string{"noclicks"}

		m.Headers["X-MC-AutoText"] = []string{"false"}
		m.Headers["X-MC-AutoHTML"] = []string{"false"}
		m.Headers["X-MC-ViewContentLink"] = []string{"false"}
	}

	client, err := smtp.Dial(net.JoinHostPort(conf.EmailSmtp.Host, strconv.Itoa(conf.EmailSmtp.Port)))
	if err != nil {
		return errors.Wrap(err, "new SMTP client")
	}
	defer func() { _ = client.Close() }()

	// NOTE: Some services (e.g. Google SMTP relay) require to echo desired hostname,
	// our current email dependency "github.com/jordan-wright/email" has no option
	// for it and always echoes "localhost" which makes it unusable.
	heloHostname := conf.EmailSmtp.Domain
	if heloHostname == "" {
		heloHostname = "localhost" // CI:LOCALHOST_OK
	}
	err = client.Hello(heloHostname)
	if err != nil {
		return errors.Wrap(err, "send HELO")
	}

	// Use TLS if available
	if ok, _ := client.Extension("STARTTLS"); ok {
		err = client.StartTLS(
			&tls.Config{
				InsecureSkipVerify: conf.EmailSmtp.NoVerifyTLS,
				ServerName:         conf.EmailSmtp.Host,
			},
		)
		if err != nil {
			return errors.Wrap(err, "send STARTTLS")
		}
	}

	var smtpAuth smtp.Auth
	switch conf.EmailSmtp.Authentication {
	case "none": // nothing to do
	case "PLAIN":
		smtpAuth = smtp.PlainAuth("", conf.EmailSmtp.Username, conf.EmailSmtp.Password, conf.EmailSmtp.Host)
	case "CRAM-MD5":
		smtpAuth = smtp.CRAMMD5Auth(conf.EmailSmtp.Username, conf.EmailSmtp.Password)
	default:
		return errors.Errorf("invalid SMTP authentication type %q", conf.EmailSmtp.Authentication)
	}

	if smtpAuth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(smtpAuth); err != nil {
				return errors.Wrap(err, "auth")
			}
		}
	}

	err = client.Mail(conf.EmailAddress)
	if err != nil {
		return errors.Wrap(err, "send MAIL")
	}
	for _, addr := range m.To {
		if err = client.Rcpt(addr); err != nil {
			return errors.Wrap(err, "send RCPT")
		}
	}
	w, err := client.Data()
	if err != nil {
		return errors.Wrap(err, "send DATA")
	}

	_, err = w.Write(raw)
	if err != nil {
		return errors.Wrap(err, "write")
	}
	err = w.Close()
	if err != nil {
		return errors.Wrap(err, "close")
	}

	err = client.Quit()
	if err != nil {
		return errors.Wrap(err, "send QUIT")
	}
	return nil
}

// MockSend is used in tests to mock the Send func.
var MockSend func(ctx context.Context, message Message) error

var disableSilently bool

// DisableSilently prevents sending of emails, even if email sending is
// configured. Use it in tests to ensure that they do not send real emails.
func DisableSilently() {
	disableSilently = true
}
