package types

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

type CodeHostConnection interface {
	// SetURN updates the URN field of the underlying connection.
	SetURN(string)
}

var _ CodeHostConnection = (*AWSCodeCommitConnection)(nil)

type AWSCodeCommitConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.AWSCodeCommitConnection
}

func (c *AWSCodeCommitConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*BitbucketCloudConnection)(nil)

type BitbucketCloudConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.BitbucketCloudConnection
}

func (c *BitbucketCloudConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*BitbucketServerConnection)(nil)

type BitbucketServerConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.BitbucketServerConnection
}

func (c *BitbucketServerConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*GitHubConnection)(nil)

type GitHubConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitHubConnection
}

func (c *GitHubConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*GitLabConnection)(nil)

type GitLabConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitLabConnection
}

func (c *GitLabConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*GitoliteConnection)(nil)

type GitoliteConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitoliteConnection
}

func (c *GitoliteConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*OtherExternalServiceConnection)(nil)

type OtherExternalServiceConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.OtherExternalServiceConnection
}

func (c *OtherExternalServiceConnection) SetURN(urn string) {
	c.URN = urn
}

var _ CodeHostConnection = (*PhabricatorConnection)(nil)

type PhabricatorConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.PhabricatorConnection
}

func (c *PhabricatorConnection) SetURN(urn string) {
	c.URN = urn
}
