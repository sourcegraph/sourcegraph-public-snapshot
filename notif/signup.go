package notif

import (
	"fmt"
	"log"
)

// SendNewUserSignUp sends an email using its arguments. This should
// be called when new users sign up.
func SendNewUserSignUp(login string, email string, name string) error {
	if email == "" {
		return fmt.Errorf("skipping welcome email for %s (email is empty)", login)
	}

	log.Printf("Sending welcome email to %s.", email)
	resp, err := SendMandrillTemplate("welcome-email", name, email, nil)
	if err != nil {
		return fmt.Errorf("error sending welcome email: %s, response was %+v", err, resp)
	}
	return nil
}
