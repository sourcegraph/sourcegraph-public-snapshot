// Package txemail sends transactional emails.
package txemail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
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
}, []string{"success", "email_source"})

// render returns the rendered message contents without sending email.
func render(fromAddress, fromName string, message Message) (*email.Email, error) {
	m := email.Email{
		To: message.To,
		From: (&mail.Address{
			Name:    fromName,
			Address: fromAddress,
		}).String(),
		Headers: make(textproto.MIMEHeader),
	}
	if message.ReplyTo != nil {
		m.ReplyTo = []string{*message.ReplyTo}
	}
	if message.MessageID != nil {
		m.Headers["Message-ID"] = []string{*message.MessageID}
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

// Send sends a transactional email if SMTP is configured. All services within the frontend
// should use this directly to send emails.  Source is used to categorize metrics, and
// should indicate the product feature that is sending this email.
//
// Callers that do not live in the frontend should call internalapi.Client.SendEmail
// instead.
//
// 🚨 SECURITY: If the email address is associated with a user, make sure to assess whether
// the email should be verified or not, and conduct the appropriate checks before sending.
// This helps reduce the chance that we damage email sender reputations when attempting to
// send emails to nonexistent email addresses.
func Send(ctx context.Context, source string, message Message) (err error) {
	if MockSend != nil {
		return MockSend(ctx, message)
	}
	if disableSilently {
		return nil
	}

	config := conf.Get()
	if config.EmailAddress == "" {
		return errors.New("no \"From\" email address configured (in email.address)")
	}
	if config.EmailSmtp == nil {
		return errors.New("no SMTP server configured (in email.smtp)")
	}

	// Previous errors are configuration errors, do not track as error. Subsequent errors
	// are delivery errors.
	defer func() {
		emailSendCounter.WithLabelValues(strconv.FormatBool(err == nil), source).Inc()
	}()

	m, err := render(config.EmailAddress, conf.EmailSenderName(), message)
	if err != nil {
		return errors.Wrap(err, "render")
	}

	// Disable Mandrill features, because they make the emails look sketchy.
	if config.EmailSmtp.Host == "smtp.mandrillapp.com" {
		// Disable click tracking ("noclicks" could be any string; the docs say that anything will disable click tracking except
		// those defined at
		// https://mandrill.zendesk.com/hc/en-us/articles/205582117-How-to-Use-SMTP-Headers-to-Customize-Your-Messages#enable-open-and-click-tracking).
		m.Headers["X-MC-Track"] = []string{"noclicks"}

		m.Headers["X-MC-AutoText"] = []string{"false"}
		m.Headers["X-MC-AutoHTML"] = []string{"false"}
		m.Headers["X-MC-ViewContentLink"] = []string{"false"}
	}

	// Apply header configuration to message
	for _, header := range config.EmailSmtp.AdditionalHeaders {
		m.Headers.Add(header.Key, header.Value)
	}

	// Generate message data
	raw, err := m.Bytes()
	if err != nil {
		return errors.Wrap(err, "get bytes")
	}

	// Set up client
	client, err := smtp.Dial(net.JoinHostPort(config.EmailSmtp.Host, strconv.Itoa(config.EmailSmtp.Port)))
	if err != nil {
		return errors.Wrap(err, "new SMTP client")
	}
	defer func() { _ = client.Close() }()

	// NOTE: Some services (e.g. Google SMTP relay) require to echo desired hostname,
	// our current email dependency "github.com/jordan-wright/email" has no option
	// for it and always echoes "localhost" which makes it unusable.
	heloHostname := config.EmailSmtp.Domain
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
				InsecureSkipVerify: config.EmailSmtp.NoVerifyTLS,
				ServerName:         config.EmailSmtp.Host,
			},
		)
		if err != nil {
			return errors.Wrap(err, "send STARTTLS")
		}
	}

	var smtpAuth smtp.Auth
	switch config.EmailSmtp.Authentication {
	case "none": // nothing to do
	case "PLAIN":
		smtpAuth = smtp.PlainAuth("", config.EmailSmtp.Username, config.EmailSmtp.Password, config.EmailSmtp.Host)
	case "CRAM-MD5":
		smtpAuth = smtp.CRAMMD5Auth(config.EmailSmtp.Username, config.EmailSmtp.Password)
	default:
		return errors.Errorf("invalid SMTP authentication type %q", config.EmailSmtp.Authentication)
	}

	if smtpAuth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(smtpAuth); err != nil {
				return errors.Wrap(err, "auth")
			}
		}
	}

	err = client.Mail(config.EmailAddress)
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
