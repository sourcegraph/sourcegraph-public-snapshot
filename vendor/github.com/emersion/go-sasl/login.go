package sasl

// The LOGIN mechanism name.
const Login = "LOGIN"

// Authenticates users with an username and a password.
type LoginAuthenticator func(username, password string) error

type loginState int

const (
	loginNotStarted loginState = iota
	loginWaitingUsername
	loginWaitingPassword
	loginCompleted
)

type loginServer struct {
	state loginState
	username, password string
	authenticate LoginAuthenticator
}

// A server implementation of the LOGIN authentication mechanism, as described
// in https://tools.ietf.org/html/draft-murchison-sasl-login-00.
func NewLoginServer(authenticator LoginAuthenticator) Server {
	return &loginServer{authenticate: authenticator}
}

func (a *loginServer) Next(response []byte) (challenge []byte, done bool, err error) {
	switch a.state {
	case loginNotStarted:
		challenge = []byte("Username:")
	case loginWaitingUsername:
		a.username = string(response)
		challenge = []byte("Password:")
	case loginWaitingPassword:
		a.password = string(response)
		err = a.authenticate(a.username, a.password)
		done = true
	default:
		err = ErrUnexpectedClientResponse
	}

	a.state++
	return
}
