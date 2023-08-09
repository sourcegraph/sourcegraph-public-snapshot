// Trimmed down platform context object to make dev and prod builds work. The
// real platform context module imports tracking/eventLogger.ts, which causes a
// couple of errors.
//

import { from } from 'rxjs'
import { map, publishReplay, refCount, shareReplay } from 'rxjs/operators'

import { createAggregateError } from '$lib/common'
import type { ViewerSettingsResult, ViewerSettingsVariables } from '$lib/graphql/shared'
import { fromObservableQuery } from '$lib/http-client'
import { type PlatformContext, gqlToCascade } from '$lib/shared'
import { viewerSettingsQuery } from '$lib/user/api/settings'

import { query, type GraphQLClient, gql } from './graphql'

export function createPlatformContext(client: GraphQLClient): Pick<PlatformContext, 'requestGraphQL' | 'settings'> {
    const settingsQueryWatcher = client.watchQuery<ViewerSettingsResult, ViewerSettingsVariables>({
        query: gql(viewerSettingsQuery),
    })

    return {
        settings: fromObservableQuery(settingsQueryWatcher).pipe(
            map(({ data, errors }) => {
                if (!data?.viewerSettings) {
                    throw createAggregateError(errors)
                }

                return data.viewerSettings
            }),
            shareReplay(1),
            map(gqlToCascade),
            publishReplay(1),
            refCount()
        ),
        requestGraphQL<R, V extends { [key: string]: any } = object>({
            request,
            variables,
        }: {
            request: string
            variables: V
        }) {
            return from(
                query<R, V>(gql(request), variables).then(
                    data => ({ data, errors: [] }),
                    error => ({ data: null, errors: [error] })
                )
            )
        },
    }
}
