import { browser } from '$app/environment'
import { fetchEvaluatedFeatureFlags } from '$lib/featureflags'
import type { CurrentAuthStateResult } from '$lib/graphql/shared'
import { getDocumentNode } from '$lib/http-client'
import { currentAuthStateQuery } from '$lib/loader/auth'
import { fetchUserSettings } from '$lib/user/api/settings'
import { getWebGraphQLClient } from '$lib/web'

import type { LayoutLoad } from './$types'

// Disable server side rendering for the whole app
export const ssr = false
export const prerender = false

if (browser) {
    // Necessary to make authenticated GrqphQL requests work
    // No idea why TS picks up Mocha.SuiteFunction for this
    window.context = {
        xhrHeaders: {
            'X-Requested-With': 'Sourcegraph',
        },
    }
}

export const load: LayoutLoad = () => {
    const graphqlClient = getWebGraphQLClient()

    return {
        graphqlClient,
        user: graphqlClient
            .then(client => client.query<CurrentAuthStateResult>({ query: getDocumentNode(currentAuthStateQuery) }))
            .then(result => result.data.currentUser),
        // Initial user settings
        settings: graphqlClient.then(fetchUserSettings),
        featureFlags: graphqlClient.then(fetchEvaluatedFeatureFlags),
    }
}
