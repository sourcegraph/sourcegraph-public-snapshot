package papertrail

import (
	"fmt"
	"log"
)

func ExampleSearch() {
	token, err := ReadToken()
	if err == ErrNoTokenFound {
		fmt.Println("ExampleSearch requires a valid Papertrail API token (which you can obtain from https://papertrailapp.com/user/edit) to be set in the PAPERTRAIL_API_TOKEN environment variable or in ~/.papertrail.yml (in the format `token: MYTOKEN`). None found; skipping.")
		return
	} else if err != nil {
		log.Fatal(err)
	}

	c := NewClient((&TokenTransport{Token: token}).Client())
	_ = c
	// output:
}
