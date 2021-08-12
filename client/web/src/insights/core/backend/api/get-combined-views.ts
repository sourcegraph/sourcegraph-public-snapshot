import { combineLatest, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

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
