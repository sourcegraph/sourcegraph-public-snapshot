package userpasswd

//go:generate ../../../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd -o mocks.go -i LockoutStore
