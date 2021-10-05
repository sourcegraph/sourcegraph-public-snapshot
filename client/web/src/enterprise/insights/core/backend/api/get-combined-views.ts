import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { ViewInsightProviderResult } from '../types'
import { createExtensionInsight } from '../utils/create-extension-insight'

/**
 * Get combined (backend and extension) code insights unified method.
 * Used for fetching insights in different places (home (search) page, directory pages)
 */
export const getCombinedViews = (
    getExtensionsInsights: () => Observable<ViewProviderResult[]>
): Observable<ViewInsightProviderResult[]> =>
    getExtensionsInsights().pipe(map(extensionInsights => extensionInsights.map(createExtensionInsight)))
