package gitolite

import (
	context "context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/gitolite/mock_gitolite"
	"github.com/sourcegraph/sourcegraph/pkg/tst"
)

type mockedClientParams struct {
	Host string

	commandOut string
}

func (p mockedClientParams) newClient(ctrl *gomock.Controller) *Client {
	return &Client{
		Host: p.Host,
		command: func() Command {
			m := mock_gitolite.NewMockCommand(ctrl)
			m.
				EXPECT().
				Output(gomock.Any(), gomock.Eq("ssh"), gomock.Eq(p.Host), gomock.Eq("info")).
				Return([]byte(p.commandOut), nil)
			return m
		}(),
	}
}

func Test_Client_ListRepos(t *testing.T) {
	tests := []struct {
		client func(ctrl *gomock.Controller) *Client

		expRepos []*Repo
		expErr   error
	}{
		{
			client: mockedClientParams{
				Host: "git@gitolite.example.com",
				commandOut: `hello admin, this is git@gitolite-799486b5db-ghrxg running gitolite3 v3.6.6-0-g908f8c6 on git 2.7.4

		 R W    gitolite-admin
		 R W    repowith@sign
		 R W    testing
		`,
			}.newClient,
			expRepos: []*Repo{
				{Name: "gitolite-admin", URL: "git@gitolite.example.com:gitolite-admin"},
				{Name: "repowith@sign", URL: "git@gitolite.example.com:repowith@sign"},
				{Name: "testing", URL: "git@gitolite.example.com:testing"},
			},
		},
		{
			client: mockedClientParams{
				Host:       "git@gitolite.example.com",
				commandOut: "",
			}.newClient,
			expRepos: nil,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := test.client(ctrl)

			repos, err := client.ListRepos(context.Background())
			tst.CheckDeepEqual(t, test.expRepos, repos, "returned repos did not match")
			tst.CheckDeepEqual(t, test.expErr, err, "returned error did not match")
		})
	}
}
