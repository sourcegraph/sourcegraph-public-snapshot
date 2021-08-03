import { Remote } from 'comlink'
import { combineLatest, from, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { ViewInsightProviderResult } from '../types'
import { createExtensionInsight } from '../utils/create-extension-insight'

import { getBackendInsights } from './get-backend-insights'

/**
 * Get combined (backend and extension) code insights unified method.
 * Used for fetching insights in different places (home (search) page, directory pages)
 */
export const getCombinedViews = (
    getExtensionsInsights: () => Observable<ViewProviderResult[]>,
    insightIds?: string[]
): Observable<ViewInsightProviderResult[]> =>
    combineLatest([
        getBackendInsights(insightIds),
        getExtensionsInsights().pipe(map(extensionInsights => extensionInsights.map(createExtensionInsight))),
    ]).pipe(map(([backendInsights, extensionViews]) => [...backendInsights, ...extensionViews]))

export const getInsightCombinedViews = (
    extensionApi: Promise<Remote<FlatExtensionHostAPI>>,
    allInsightIds?: string[],
    backendInsightIds?: string[]
): Observable<ViewInsightProviderResult[]> =>
    getCombinedViews(
        () =>
            from(extensionApi).pipe(
                switchMap(extensionHostAPI =>
                    wrapRemoteObservable(extensionHostAPI.getInsightsViews({}, allInsightIds))
                )
            ),
        backendInsightIds
    )
