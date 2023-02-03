package azureoauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/gologin"
	oauth2Login "github.com/dghubble/gologin/oauth2"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	extsvcauth "github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init(logger log.Logger, db database.DB) {
	const pkgName = "azuredoauth"
	logger = logger.Scoped(pkgName, "Azure DevOps OAuth config watch")
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := parseConfig(logger, cfg, db)
		return problems
	})

	go func() {
		conf.Watch(func() {
			newProviders, _ := parseConfig(logger, conf.Get(), db)
			if len(newProviders) == 0 {
				providers.Update(pkgName, nil)
				return
			}

			if err := licensing.Check(licensing.FeatureSSO); err != nil {
				logger.Error("Check license for SSO (Azure DevOps OAuth)", log.Error(err))
				providers.Update(pkgName, nil)
				return
			}

			newProvidersList := make([]providers.Provider, 0, len(newProviders))
			for _, p := range newProviders {
				newProvidersList = append(newProvidersList, p.Provider)
			}
			providers.Update(pkgName, newProvidersList)
		})
	}()
}

type Provider struct {
	*schema.AzureDevOpsAuthProvider
	providers.Provider
}

func parseConfig(logger log.Logger, cfg conftypes.SiteConfigQuerier, db database.DB) (ps []Provider, problems conf.Problems) {
	// TODO: Abstract away in a function maybe for DRY.
	externalURL, err := url.Parse(cfg.SiteConfig().ExternalURL)
	if err != nil {
		problems = append(problems, conf.NewSiteProblem("Could not parse `externalURL`, which is needed to determine the OAuth callback URL."))

		return ps, problems
	}

	callbackURL := *externalURL
	callbackURL.Path = "/.auth/azuredevops/callback"

	for _, pr := range cfg.SiteConfig().AuthProviders {
		if pr.AzureDevOps == nil {
			continue
		}

		// FIXME: If we're passing pr and pr.AzureDevOps, might as well just pass pr.
		provider, providerProblems := parseProvider(logger, pr.AzureDevOps, db, pr, callbackURL)
		problems = append(problems, conf.NewSiteProblems(providerProblems...)...)

		if provider == nil {
			continue
		}

		ps = append(ps, Provider{
			AzureDevOpsAuthProvider: pr.AzureDevOps,
			Provider:                provider,
		})
	}

	return ps, problems
}

func parseProvider(logger log.Logger, p *schema.AzureDevOpsAuthProvider, db database.DB, sourceCfg schema.AuthProviders, callbackURL url.URL) (provider *oauth.Provider, messages []string) {
	// TODO: Handle empty p.Url. Or do I need to? I have a default?
	// app url is vscode
	parsedURL, err := url.Parse(p.Url)
	if err != nil {
		messages = append(messages, fmt.Sprintf("Failed to parse Azure DevOps URL %q. Login via this Azure instance will not work.", p.Url))
		return nil, messages
	}

	codeHost := extsvc.NewCodeHost(parsedURL, extsvc.KindAzureDevOps)

	// TODO: App secret vs client secret
	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix: authPrefix,
		OAuth2Config: func() oauth2.Config {
			return oauth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Scopes:       strings.Split(p.ApiScope, ","),
				Endpoint: oauth2.Endpoint{
					AuthURL:   "https://app.vssps.visualstudio.com/oauth2/authorize",
					TokenURL:  "https://app.vssps.visualstudio.com/oauth2/token",
					AuthStyle: oauth2.AuthStyleInParams,
				},
				RedirectURL: callbackURL.String(),
			}
		},
		SourceConfig: sourceCfg,
		// TODO: Use this util in other places where we have the function getStateConfig.
		StateConfig: oauth.GetStateConfig("azure-state-cookie"),
		ServiceID:   parsedURL.String(),
		ServiceType: extsvc.TypeAzureDevOps,
		Login:       loginHandler,
		Callback: func(config oauth2.Config) http.Handler {
			return callbackHandler(
				logger,
				&config,
				oauth.SessionIssuer(logger, db, &sessionIssuerHelper{
					db:          db,
					CodeHost:    codeHost,
					clientID:    p.ClientID,
					allowSignup: p.AllowSignup,
				}, sessionKey),
			)
		},
	}), messages
}

const (
	errorKey key = iota
)

func ErrorFromContext(ctx context.Context) error {
	err, ok := ctx.Value(errorKey).(error)
	if !ok {
		return fmt.Errorf("Context missing error value")
	}
	return err
}

func failureHandler(w http.ResponseWriter, req *http.Request) {
	l := log.Scoped("azuredevops.failureHandler", "failureHandler")
	l.Warn("here")

	ctx := req.Context()
	err := ErrorFromContext(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// should be unreachable, ErrorFromContext always returns some non-nil error
	http.Error(w, "", http.StatusBadRequest)
}

func loginHandler(c oauth2.Config) http.Handler {
	// c.RedirectURL
	l := log.Scoped("azuredevops.loginHandler", "loginHandler")
	l.Warn("here", log.String("oauth2.Config", fmt.Sprintf("%#v", c)))
	return oauth2Login.LoginHandler(&c, http.HandlerFunc(failureHandler))
}

// TODO: Better naming for callbackHandler vs CallbackHandler? Surely this can be simplified?
func callbackHandler(logger log.Logger, config *oauth2.Config, success http.Handler) http.Handler {
	l := log.Scoped("azuredevops.callbackHandler", "callbackHandler")
	l.Warn("here", log.String("oauth2.Config", fmt.Sprintf("%#v", config)))

	success = azureDevOpsHandler(logger, config, success, http.HandlerFunc(failureHandler))

	return CallbackHandler(config, success, gologin.DefaultFailureHandler)
}

func azureDevOpsHandler(logger log.Logger, config *oauth2.Config, success, failure http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {

		l := log.Scoped("azureDevOpsSuccessHandler", "azure dev ops handler")
		l.Warn("here", log.String("oauth2.Config", fmt.Sprintf("%#v", config)))

		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)

		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		l.Warn("token", log.String("oauth2 token", token.AccessToken))

		// TODO: Finish implementation
		azureClient, err := azureDevOpsClientFromAuthURL(config.Endpoint.AuthURL, token.AccessToken)
		if err != nil {
			logger.Error("failed to create azuredevops.Client", log.String("error", err.Error()))
			ctx = gologin.WithError(ctx, errors.Errorf("failed to create HTTP client for azuredevops with AuthURL %q", config.Endpoint.AuthURL))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// TODO: Probably don't need this
		profile, err := azureClient.AzureServicesProfile(ctx)
		if err != nil {
			msg := "failed to get Azure profile after oauth2 callback"
			logger.Error(msg, log.String("error", err.Error()))
			ctx = gologin.WithError(ctx, errors.Wrap(err, msg))
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		if profile.ID == "" || profile.EmailAddress == "" {
			msg := "bad Azure profile in API response"
			logger.Error(msg, log.String("profile", fmt.Sprintf("%#v", profile)))

			ctx = gologin.WithError(
				ctx,
				errors.Errorf(fmt.Sprintf("%s: %#v", msg, profile)),
			)

			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// FIXME: Implement this.
		// err = validateResponse(user, err)
		// if err != nil {
		// 	// TODO: Copy pasta
		// 	// TODO: Prefer a more general purpose fix, potentially
		// 	// https://github.com/sourcegraph/sourcegraph/pull/20000
		// 	logger.Warn("invalid response", log.Error(err))
		// }
		// if err != nil {
		// 	ctx = gologin.WithError(ctx, err)
		// 	failure.ServeHTTP(w, req.WithContext(ctx))
		// 	return
		// }

		ctx = withUser(ctx, profile)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func azureDevOpsClientFromAuthURL(authURL, oauthToken string) (*azuredevops.Client, error) {
	baseURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}
	baseURL.Path = ""
	baseURL.RawQuery = ""
	baseURL.Fragment = ""

	httpCli, err := httpcli.ExternalClientFactory.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http client, this is likely a misconfiguration")
	}

	cli, err := azuredevops.NewClient(
		urnAzureDevOpsOAuth,
		baseURL,
		&extsvcauth.OAuthBearerToken{Token: oauthToken},
		httpCli,
	)

	return cli, nil
}

const authPrefix = auth.AuthURLPrefix + "/azuredevops"
const sessionKey = "azuredevopsoauth@0"
const urnAzureDevOpsOAuth = "AzureDevOpsOAuth"
