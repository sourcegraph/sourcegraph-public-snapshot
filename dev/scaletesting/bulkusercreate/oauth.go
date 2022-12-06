package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/google/go-github/v41/github"

	"github.com/sourcegraph/sourcegraph/lib/group"
)

// generateUserOAuthCsv creates user impersonation OAuth tokens and writes them to a CSV file together with the usernames.
func generateUserOAuthCsv(ctx context.Context, users []*user, tokensDone int64) {
	tg := group.NewWithResults[userToken]().WithMaxConcurrency(1000)
	for _, u := range users {
		currentU := u
		tg.Go(func() userToken {
			token := executeCreateUserImpersonationToken(ctx, currentU)
			atomic.AddInt64(&tokensDone, 1)
			progress.SetValue(5, float64(tokensDone))
			return userToken{
				login: currentU.Login,
				token: token,
			}
		})
	}
	pairs := tg.Wait()

	csvFile, err := os.Create("users.csv")
	defer csvFile.Close()
	if err != nil {
		log.Fatalf("Failed creating csv: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	_ = csvwriter.Write([]string{"login", "token"})

	// sort by username
	sort.Slice(pairs, func(i, j int) bool {
		comp := strings.Compare(pairs[i].login, pairs[j].login)
		return comp == -1
	})

	for _, pair := range pairs {
		if err = csvwriter.Write([]string{pair.login, pair.token}); err != nil {
			log.Fatalln("error writing pair to file", err)
		}
	}
}

// executeCreateUserImpersonationToken creates a user impersonation OAuth token for the given user.
func executeCreateUserImpersonationToken(ctx context.Context, u *user) string {
	auth, _, err := gh.Admin.CreateUserImpersonation(ctx, u.Login, &github.ImpersonateUserOptions{Scopes: []string{"repo", "read:org", "read:user_email"}})
	if err != nil {
		log.Fatal(err)
	}

	return auth.GetToken()
}
