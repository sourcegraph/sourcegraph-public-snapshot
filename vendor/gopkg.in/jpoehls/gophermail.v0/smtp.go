package gophermail

import (
    "crypto/tls"
	"net/smtp"
)

// SendMail connects to the server at addr, switches to TLS if possible,
// authenticates with mechanism a if possible, and then sends the given Message.
//
// Based heavily on smtp.SendMail().
func SendMail(addr string, a smtp.Auth, msg *Message) error {
	msgBytes, err := msg.Bytes()
	if err != nil {
		return err
	}

	var to []string
	for _, address := range msg.To {
		to = append(to, address.Address)
	}

	for _, address := range msg.Cc {
		to = append(to, address.Address)
	}

	for _, address := range msg.Bcc {
		to = append(to, address.Address)
	}

	return smtp.SendMail(addr, a, msg.From.Address, to, msgBytes)
}


// SendTLSMail does the same thing as SendMail, except with the added
// option of providing a tls.Config
func SendTLSMail(addr string, a smtp.Auth, msg *Message, cfg tls.Config) error {
	msgBytes, err := msg.Bytes()
	if err != nil {
		return err
	}

	var to []string
	for _, address := range msg.To {
		to = append(to, address.Address)
	}

	for _, address := range msg.Cc {
		to = append(to, address.Address)
	}

	for _, address := range msg.Bcc {
		to = append(to, address.Address)
	}

    from := msg.From.String()

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
        if err = c.StartTLS(&cfg); err != nil {
            return err
        }
	}

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msgBytes)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
