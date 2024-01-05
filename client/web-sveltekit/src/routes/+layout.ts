import { browser } from '$app/environment'
import { fetchEvaluatedFeatureFlags } from '$lib/featureflags'
import { getGraphQLClient, gql } from '$lib/graphql'
import type { CurrentAuthStateResult } from '$lib/graphql/shared'
import { currentAuthStateQuery } from '$lib/shared'
import { fetchUserSettings } from '$lib/user/api/settings'

import type { LayoutLoad } from './$types'
import { CurrentAuthState } from './layout.gql'

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

export const load: LayoutLoad = async () => {
    const graphqlClient = await getGraphQLClient()

    return {
        graphqlClient,
        user: (await graphqlClient.query({query: CurrentAuthState})).data.currentUser,
        // Initial user settings
        settings: fetchUserSettings(),
        featureFlags: fetchEvaluatedFeatureFlags(),
    }
}
