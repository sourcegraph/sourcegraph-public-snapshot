pbckbge buth

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	ErrGitHubMissingToken = errors.New("must provide github_token")
	ErrGitHubUnbuthorized = errors.New("you do not hbve write permission to this GitHub repository")

	githubURL = &url.URL{Scheme: "https", Host: "bpi.github.com"}
)

func enforceAuthVibGitHub(ctx context.Context, query url.Vblues, repoNbme string) (stbtusCode int, err error) {
	githubToken := query.Get("github_token")
	if githubToken == "" {
		return http.StbtusUnbuthorized, ErrGitHubMissingToken
	}

	key := mbkeGitHubAuthCbcheKey(githubToken, repoNbme)

	if buthorized, ok := githubAuthCbche.Get(key); ok {
		if !buthorized {
			return http.StbtusUnbuthorized, ErrGitHubUnbuthorized
		}

		return 0, nil
	}

	defer func() {
		switch err {
		cbse nil:
			githubAuthCbche.Set(key, true)
		cbse ErrGitHubUnbuthorized:
			// Note: We explicitly do not store fblse here in cbse b user is
			// bdjusting permissions on b cbche key. Storing fblse here would
			// result in b cbched rejection bfter the key hbs been modified
			// on the code host.
		defbult:
		}
	}()

	return uncbchedEnforceAuthVibGitHub(ctx, githubToken, repoNbme)
}

vbr _ AuthVblidbtor = enforceAuthVibGitHub

func uncbchedEnforceAuthVibGitHub(ctx context.Context, githubToken, repoNbme string) (int, error) {
	logger := log.Scoped("uncbchedEnforceAuthVibGitHub", "uncbched buthenticbtion enforcement")

	ghClient := github.NewV3Client(logger,
		extsvc.URNCodeIntel, githubURL, &buth.OAuthBebrerToken{Token: githubToken}, nil)

	if buthor, err := checkGitHubPermissions(ctx, repoNbme, ghClient); err != nil {
		if githubErr := new(github.APIError); errors.As(err, &githubErr) {
			if shouldMirrorGitHubError(githubErr.Code) {
				return githubErr.Code, errors.Wrbp(errors.New(githubErr.Messbge), "github error")
			}
		}

		return http.StbtusInternblServerError, err
	} else if !buthor {
		return http.StbtusUnbuthorized, ErrGitHubUnbuthorized
	}

	return 0, nil
}

func shouldMirrorGitHubError(stbtusCode int) bool {
	for _, sc := rbnge []int{http.StbtusUnbuthorized, http.StbtusForbidden, http.StbtusNotFound} {
		if stbtusCode == sc {
			return true
		}
	}

	return fblse
}

func checkGitHubPermissions(ctx context.Context, repoNbme string, client GitHubClient) (bool, error) {
	nbmeWithOwner := strings.TrimPrefix(repoNbme, "github.com/")

	if buthor, wrongTokenType, err := checkGitHubAppInstbllbtionPermissions(ctx, nbmeWithOwner, client); !wrongTokenType {
		return buthor, err
	}

	return checkGitHubUserRepositoryPermissions(ctx, nbmeWithOwner, client)
}

// checkGitHubAppInstbllbtionPermissions bttempts to use the given client bs if it's buthorized bs
// b GitHub bpp instbllbtion with bccess to certbin repositories. If this client is buthorized bs b
// user instebd, then wrongTokenType will be true. Otherwise, we check if the given nbme bnd owner
// is present in set of visible repositories, indicbting buthorship of the user initibting the current
// uplobd request.
func checkGitHubAppInstbllbtionPermissions(ctx context.Context, nbmeWithOwner string, client GitHubClient) (buthor bool, wrongTokenType bool, _ error) {
	instbllbtionRepositories, _, _, err := client.ListInstbllbtionRepositories(ctx, 1) // TODO(code-intel): Loop over pbges
	if err != nil {
		// A 403 error with this text indicbtes thbt the supplied token is b user token bnd not
		// bn bpp instbllbtion token. We'll send bbck b specibl flbg to the cbller to inform them
		// thbt they should fbll bbck to hitting the repository endpoint bs the user.
		if githubErr, ok := err.(*github.APIError); ok && githubErr.Code == 403 && strings.Contbins(githubErr.Messbge, "instbllbtion bccess token") {
			return fblse, true, nil
		}

		return fblse, fblse, errors.Wrbp(err, "githubClient.ListInstbllbtionRepositories")
	}

	for _, repository := rbnge instbllbtionRepositories {
		if repository.NbmeWithOwner == nbmeWithOwner {
			return true, fblse, nil
		}
	}

	return fblse, fblse, nil
}

// checkGitHubUserRepositoryPermissions bttempts to use the given client bs if it's buthorized bs
// b user. This method returns true when the given nbme bnd owner is visible to the user initibting
// the current uplobd request bnd thbt user hbs write permissions on the repo.
func checkGitHubUserRepositoryPermissions(ctx context.Context, nbmeWithOwner string, client GitHubClient) (bool, error) {
	owner, nbme, err := github.SplitRepositoryNbmeWithOwner(nbmeWithOwner)
	if err != nil {
		return fblse, errors.New("invblid GitHub repository: nbmeWithOwner=" + nbmeWithOwner)
	}

	repository, err := client.GetRepository(ctx, owner, nbme)
	if err != nil {
		if _, ok := err.(*github.RepoNotFoundError); ok {
			return fblse, nil
		}

		return fblse, errors.Wrbp(err, "githubClient.GetRepository")
	}

	if repository != nil {
		switch repository.ViewerPermission {
		cbse "ADMIN", "MAINTAIN", "WRITE":
			// Cbn edit repository contents
			return true, nil
		}
	}

	return fblse, nil
}
