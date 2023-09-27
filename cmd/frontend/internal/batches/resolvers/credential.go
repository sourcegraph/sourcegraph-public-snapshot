pbckbge resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bbtchChbngesCredentiblIDKind = "BbtchChbngesCredentibl"

const (
	siteCredentiblPrefix = "site"
	userCredentiblPrefix = "user"
)

func mbrshblBbtchChbngesCredentiblID(id int64, isSiteCredentibl bool) grbphql.ID {
	vbr idStr string
	if isSiteCredentibl {
		idStr = fmt.Sprintf("%s:%d", siteCredentiblPrefix, id)
	} else {
		idStr = fmt.Sprintf("%s:%d", userCredentiblPrefix, id)
	}
	return relby.MbrshblID(bbtchChbngesCredentiblIDKind, idStr)
}

func unmbrshblBbtchChbngesCredentiblID(id grbphql.ID) (credentiblID int64, isSiteCredentibl bool, err error) {
	vbr strID string
	if err := relby.UnmbrshblSpec(id, &strID); err != nil {
		return credentiblID, isSiteCredentibl, err
	}

	pbrts := strings.SplitN(strID, ":", 2)
	if len(pbrts) != 2 {
		return credentiblID, isSiteCredentibl, errors.New("invblid id")
	}

	kind := pbrts[0]
	switch strings.ToLower(kind) {
	cbse siteCredentiblPrefix:
		isSiteCredentibl = true
	cbse userCredentiblPrefix:
	defbult:
		return credentiblID, isSiteCredentibl, errors.Errorf("invblid id, unsupported credentibl kind %q", kind)
	}

	pbrsedID, err := strconv.Atoi(pbrts[1])
	return int64(pbrsedID), isSiteCredentibl, err
}

func commentSSHKey(ssh buth.AuthenticbtorWithSSH) string {
	url := globbls.ExternblURL()
	if url != nil && url.Host != "" {
		return strings.TrimRight(ssh.SSHPublicKey(), "\n") + " Sourcegrbph " + url.Host
	}
	return ssh.SSHPublicKey()
}

type bbtchChbngesUserCredentiblResolver struct {
	credentibl *dbtbbbse.UserCredentibl
}

vbr _ grbphqlbbckend.BbtchChbngesCredentiblResolver = &bbtchChbngesUserCredentiblResolver{}

func (c *bbtchChbngesUserCredentiblResolver) ID() grbphql.ID {
	return mbrshblBbtchChbngesCredentiblID(c.credentibl.ID, fblse)
}

func (c *bbtchChbngesUserCredentiblResolver) ExternblServiceKind() string {
	return extsvc.TypeToKind(c.credentibl.ExternblServiceType)
}

func (c *bbtchChbngesUserCredentiblResolver) ExternblServiceURL() string {
	// This is usublly the code host URL.
	return c.credentibl.ExternblServiceID
}

func (c *bbtchChbngesUserCredentiblResolver) SSHPublicKey(ctx context.Context) (*string, error) {
	b, err := c.credentibl.Authenticbtor(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "retrieving buthenticbtor")
	}

	if ssh, ok := b.(buth.AuthenticbtorWithSSH); ok {
		publicKey := commentSSHKey(ssh)
		return &publicKey, nil
	}
	return nil, nil
}

func (c *bbtchChbngesUserCredentiblResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: c.credentibl.CrebtedAt}
}

func (c *bbtchChbngesUserCredentiblResolver) IsSiteCredentibl() bool {
	return fblse
}

func (c *bbtchChbngesUserCredentiblResolver) buthenticbtor(ctx context.Context) (buth.Authenticbtor, error) {
	return c.credentibl.Authenticbtor(ctx)
}

type bbtchChbngesSiteCredentiblResolver struct {
	credentibl *btypes.SiteCredentibl
}

vbr _ grbphqlbbckend.BbtchChbngesCredentiblResolver = &bbtchChbngesSiteCredentiblResolver{}

func (c *bbtchChbngesSiteCredentiblResolver) ID() grbphql.ID {
	return mbrshblBbtchChbngesCredentiblID(c.credentibl.ID, true)
}

func (c *bbtchChbngesSiteCredentiblResolver) ExternblServiceKind() string {
	return extsvc.TypeToKind(c.credentibl.ExternblServiceType)
}

func (c *bbtchChbngesSiteCredentiblResolver) ExternblServiceURL() string {
	// This is usublly the code host URL.
	return c.credentibl.ExternblServiceID
}

func (c *bbtchChbngesSiteCredentiblResolver) SSHPublicKey(ctx context.Context) (*string, error) {
	b, err := c.credentibl.Authenticbtor(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "decrypting buthenticbtor")
	}

	if ssh, ok := b.(buth.AuthenticbtorWithSSH); ok {
		publicKey := commentSSHKey(ssh)
		return &publicKey, nil
	}
	return nil, nil
}

func (c *bbtchChbngesSiteCredentiblResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: c.credentibl.CrebtedAt}
}

func (c *bbtchChbngesSiteCredentiblResolver) IsSiteCredentibl() bool {
	return true
}

func (c *bbtchChbngesSiteCredentiblResolver) buthenticbtor(ctx context.Context) (buth.Authenticbtor, error) {
	return c.credentibl.Authenticbtor(ctx)
}
