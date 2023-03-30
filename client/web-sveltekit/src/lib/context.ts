// Trimmed down platform context object to make dev and prod builds work. The
// real platform context module imports tracking/eventLogger.ts, which causes a
// couple of errors.
// TODO: Consolidate with actual platform context
//
import type { ApolloQueryResult, ObservableQuery } from '@apollo/client'
import { map, publishReplay, refCount, shareReplay } from 'rxjs/operators'

import { createAggregateError } from '$lib/common'
import type { ViewerSettingsResult, ViewerSettingsVariables } from '$lib/graphql/shared'
import { getDocumentNode, type GraphQLClient, fromObservableQuery } from '$lib/http-client'
import { viewerSettingsQuery } from '$lib/loader/settings'
import { type PlatformContext, type SettingsSubject, gqlToCascade } from '$lib/shared'
import { requestGraphQL } from '$lib/web'

export function createPlatformContext(client: GraphQLClient): Pick<PlatformContext, 'requestGraphQL' | 'settings'> {
    const settingsQueryWatcher = watchViewerSettingsQuery(client)

    return {
        settings: fromObservableQuery(settingsQueryWatcher).pipe(
            map(mapViewerSettingsResult),
            shareReplay(1),
            map(gqlToCascade),
            publishReplay(1),
            refCount()
        ),
        requestGraphQL: ({ request, variables }) => requestGraphQL(request, variables),
    }
}

function mapViewerSettingsResult({ data, errors }: ApolloQueryResult<ViewerSettingsResult>): {
    subjects: SettingsSubject[]
} {
    if (!data?.viewerSettings) {
        throw createAggregateError(errors)
    }

    return data.viewerSettings
}

/**
 * Creates Apollo query watcher for the viewer's settings. Watcher is used instead of the one-time query because we
 * want to use cached response if it's available. Callers should use settingsRefreshes#next instead of calling
 * this function, to ensure that the result is propagated consistently throughout the app instead of only being
 * returned to the caller.
 */
function watchViewerSettingsQuery(
    graphQLClient: GraphQLClient
): ObservableQuery<ViewerSettingsResult, ViewerSettingsVariables> {
    return graphQLClient.watchQuery<ViewerSettingsResult, ViewerSettingsVariables>({
        query: getDocumentNode(viewerSettingsQuery),
    })
}
