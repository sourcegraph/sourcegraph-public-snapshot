import { browser } from '$app/environment'
import { fetchEvaluatedFeatureFlags } from '$lib/featureflags'
import { getGraphQLClient } from '$lib/graphql'
import { fetchUserSettings } from '$lib/user/api/settings'
import { fetchAuthenticatedUser } from '$lib/user/api/user'

import type { LayoutLoad } from './$types'

// Disable server side rendering for the whole app
export const ssr = false
export const prerender = false

if (browser) {
    // Necessary to make authenticated GrqphQL requests work
    // No idea why TS picks up Mocha.SuiteFunction for this
    // @ts-ignore
    window.context = {
        xhrHeaders: {
            'X-Requested-With': 'Sourcegraph',
        },
    }
}

export const load: LayoutLoad = () => ({
    graphqlClient: getGraphQLClient(),
    user: fetchAuthenticatedUser(),
    // Initial user settings
    settings: fetchUserSettings(),
    featureFlags: fetchEvaluatedFeatureFlags(),
})
