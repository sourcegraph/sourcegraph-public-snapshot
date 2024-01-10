import { error } from '@sveltejs/kit'

import { browser } from '$app/environment'
import { isErrorLike, parseJSONCOrError } from '$lib/common'
import { getGraphQLClient } from '$lib/graphql'
import type { Settings } from '$lib/shared'

import type { LayoutLoad } from './$types'
import { Init, EvaluatedFeatureFlagsQuery } from './layout.gql'

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
    const result = await graphqlClient.query({ query: Init, fetchPolicy: 'no-cache' })

    const settings = parseJSONCOrError<Settings>(result.data.viewerSettings.final)
    if (isErrorLike(settings)) {
        throw error(500, `Failed to parse user settings: ${settings.message}`)
    }

    return {
        graphqlClient,
        user: result.data.currentUser,
        // Initial user settings
        settings,
        featureFlags: result.data.evaluatedFeatureFlags,
        fetchEvaluatedFeatureFlags: async () => {
            const result = await graphqlClient.query({ query: EvaluatedFeatureFlagsQuery, fetchPolicy: 'no-cache' })
            return result.data.evaluatedFeatureFlags
        },
    }
}
