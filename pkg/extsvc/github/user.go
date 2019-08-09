package github

import (
	"github.com/google/go-github/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"golang.org/x/oauth2"
)

func GetExternalAccountData(data *extsvc.ExternalAccountData) (usr *github.User, tok *oauth2.Token, err error) {
	var (
		u github.User
		t oauth2.Token
	)

	if data.AccountData != nil {
		if err := data.GetAccountData(&u); err != nil {
			return nil, nil, err
		}
		usr = &u
	}
	if data.AuthData != nil {
		if err := data.GetAuthData(&t); err != nil {
			return nil, nil, err
		}
		tok = &t
	}
	return usr, tok, nil
}

func SetExternalAccountData(data *extsvc.ExternalAccountData, user *github.User, token *oauth2.Token) {
	data.SetAccountData(user)
	data.SetAuthData(token)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_799(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
