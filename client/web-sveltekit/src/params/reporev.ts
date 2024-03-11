import type { ParamMatcher } from '@sveltejs/kit'

// These are top level paths that are Sourcegraph pages, not repositories.
// By explicitly excluding then we force SvelteKit to _not_ match them, which
// will cause SvelteKit to fetch the page from the server, which will then
// load the React version.
// Note that any routes in the `routes/` directory will be handled by the
// SvelteKit app, whether they are in this list or not.
// This list is taken from 'cmd/frontend/internal/app/ui/router.go'.
const topLevelPaths = [
    'insights',
    'search-jobs',
    'setup',
    'batch-changes',
    'code-monitoring',
    'notebooks',
    'request-access',
    'api/console',
    'sign-in',
    'ping-from-self-hosted',
    'sign-up',
    'threads',
    'organizations',
    'teams',
    'settings',
    'site-admin',
    'snippets',
    'subscriptions',
    'views',
    'own',
    'contexts',
    'registry',
    'search/cody',
    'app',
    'cody',
    'get-cody',
    'post-sign-up',
    'unlock-account',
    'password-reset',
    'survey',
    'welcome',
    'embed',
    'users',
    'user',
    'search',

    // sourcegraph.com specific routes that redirect to subdomains
    // are ignored (for now)

    // community search contexts
    'kubernetes',
    'stanford',
    'stackstorm',
    'temporal',
    'o3de',
    'chakraui',
    'julia',
    'backstage',

    // legacy routes
    'login',
    'careers',
    'extensions',

    // Help pages
    'help',
]

const topLevelPathRegex = new RegExp(`^(${topLevelPaths.join('|')})($|/)`)

// This ensures that we never consider paths containing /-/ and pointing
// to non-existing pages as repo name
export const match: ParamMatcher = param => {
    // Note: param doesn't have a leading slash
    return !topLevelPathRegex.test(param) && !param.includes('/-/')
}
