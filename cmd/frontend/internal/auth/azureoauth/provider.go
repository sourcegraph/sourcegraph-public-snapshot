package azureoauth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"github.com/dghubble/gologin/v2"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/oauth"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	authPrefix = auth.AuthURLPrefix + "/azuredevops"
	sessionKey = "azuredevopsoauth@0"
)

func Init(logger log.Logger, db database.DB) {
	const pkgName = "azureoauth"
	logger = logger.Scoped(pkgName)
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := parseConfig(logger, cfg, db)
		return problems
	})

	go conf.Watch(func() {
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
}

type Provider struct {
	*schema.AzureDevOpsAuthProvider
	providers.Provider
}

func parseConfig(logger log.Logger, cfg conftypes.SiteConfigQuerier, db database.DB) (ps []Provider, problems conf.Problems) {
	callbackURL, err := azuredevops.GetRedirectURL(cfg)
	if err != nil {
		problems = append(problems, conf.NewSiteProblem(err.Error()))
		return ps, problems
	}

	existingProviders := make(collections.Set[string])

	for _, pr := range cfg.SiteConfig().AuthProviders {
		if pr.AzureDevOps == nil {
			continue
		}

		setProviderDefaults(pr.AzureDevOps)

		provider, providerProblems := parseProvider(logger, db, pr, *callbackURL)
		problems = append(problems, conf.NewSiteProblems(providerProblems...)...)

		if provider == nil {
			continue
		}

		if existingProviders.Has(provider.CachedInfo().UniqueID()) {
			problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Cannot have more than one auth provider for Azure Dev Ops with Client ID %q, only the first one will be used", pr.AzureDevOps.ClientID)))
			continue
		}

		ps = append(ps, Provider{
			AzureDevOpsAuthProvider: pr.AzureDevOps,
			Provider:                provider,
		})

		existingProviders.Add(provider.CachedInfo().UniqueID())
	}

	return ps, problems
}

// setProviderDefaults will mutate the AzureDevOpsAuthProvider with default values from the schema
// if they are not set in the config.
func setProviderDefaults(p *schema.AzureDevOpsAuthProvider) {
	if p.ApiScope == "" {
		p.ApiScope = "vso.code,vso.identity,vso.project"
	}
}

func parseProvider(logger log.Logger, db database.DB, sourceCfg schema.AuthProviders, callbackURL url.URL) (provider *oauth.Provider, messages []string) {
	// The only call site of parseProvider is parseConfig where we already check for a nil Azure
	// auth provider. But adding the check here guards against future bugs.
	if sourceCfg.AzureDevOps == nil {
		messages = append(messages, "Cannot parse nil AzureDevOps provider (this is likely a bug in the invocation of parseProvider)")
		return nil, messages
	}

	azureProvider := sourceCfg.AzureDevOps

	// Since this provider is for dev.azure.com only, we can hardcode the provider's URL to
	// azuredevops.VisualStudioAppURL.
	parsedURL, err := url.Parse(azuredevops.VisualStudioAppURL)
	if err != nil {
		messages = append(messages, fmt.Sprintf(
			"Failed to parse Azure DevOps URL %q. Login via this Azure instance will not work.", azuredevops.VisualStudioAppURL,
		))
		return nil, messages
	}

	codeHost := extsvc.NewCodeHost(parsedURL, extsvc.TypeAzureDevOps)

	allowedOrgs := map[string]struct{}{}
	for _, org := range azureProvider.AllowOrgs {
		allowedOrgs[org] = struct{}{}
	}

	sessionHandler := oauth.SessionIssuer(
		logger,
		db,
		&sessionIssuerHelper{
			db:          db,
			CodeHost:    codeHost,
			clientID:    azureProvider.ClientID,
			allowOrgs:   allowedOrgs,
			allowSignup: azureProvider.AllowSignup,
		},
		sessionKey,
	)

	authURL, err := url.JoinPath(azuredevops.VisualStudioAppURL, "/oauth2/authorize")
	if err != nil {
		messages = append(messages, fmt.Sprintf(
			"Failed to generate auth URL (this is likely a misconfigured URL in the constant azuredevops.VisualStudioAppURL): %s",
			err.Error(),
		))
		return nil, messages
	}

	tokenURL, err := url.JoinPath(azuredevops.VisualStudioAppURL, "/oauth2/token")
	if err != nil {
		messages = append(messages, fmt.Sprintf(
			"Failed to generate token URL (this is likely a misconfigured URL in the constant azuredevops.VisualStudioAppURL): %s", err.Error(),
		))
		return nil, messages
	}

	return oauth.NewProvider(oauth.ProviderOp{
		AuthPrefix: authPrefix,
		OAuth2Config: func() oauth2.Config {
			return oauth2.Config{
				ClientID:     azureProvider.ClientID,
				ClientSecret: azureProvider.ClientSecret,
				Scopes:       strings.Split(azureProvider.ApiScope, ","),
				Endpoint: oauth2.Endpoint{
					AuthURL:  authURL,
					TokenURL: tokenURL,
					// The access_token request wants the body as application/x-www-form-urlencoded. See:
					// https://learn.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#http-request-body---authorize-app
					AuthStyle: oauth2.AuthStyleInParams,
				},
				RedirectURL: callbackURL.String(),
			}
		},
		SourceConfig: sourceCfg,
		ServiceID:    azuredevops.AzureDevOpsAPIURL,
		ServiceType:  extsvc.TypeAzureDevOps,
		Login:        loginHandler,
		Callback: func(config oauth2.Config) http.Handler {
			success := azureDevOpsHandler(logger, &config, sessionHandler, gologin.DefaultFailureHandler)

			return callbackHandler(&config, success)
		},
	}), messages
}
