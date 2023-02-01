import { from } from 'rxjs'
import { map, switchMap, take } from 'rxjs/operators'

import type { LayoutLoad } from './$types'

import { browser } from '$app/environment'
import { isErrorLike } from '$lib/common'
import { createPlatformContext } from '$lib/context'
import type { CurrentAuthStateResult } from '$lib/graphql/shared'
import { getDocumentNode } from '$lib/http-client'
import { currentAuthStateQuery } from '$lib/loader/auth'
import { getWebGraphQLClient } from '$lib/web'

// Disable server side rendering for the whole app
export const ssr = false

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
    const platformContext = graphqlClient.then(createPlatformContext)

    return {
        platformContext,
        graphqlClient,
        user: graphqlClient
            .then(client => client.query<CurrentAuthStateResult>({ query: getDocumentNode(currentAuthStateQuery) }))
            .then(result => result.data.currentUser),
        // Initial user settings
        settings: from(platformContext)
            .pipe(
                switchMap(platformContext => platformContext.settings),
                map(settingsOrError => (isErrorLike(settingsOrError.final) ? null : settingsOrError.final)),
                take(1)
            )
            .toPromise(),
    }
}
