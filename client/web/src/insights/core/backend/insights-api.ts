import { Remote } from 'comlink'
import { combineLatest, from, Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { fetchBackendInsights } from './requests/fetch-backend-insights'
import { ApiService, ViewInsightProviderResult, ViewInsightProviderSourceType } from './types'
import { createViewContent } from './utils/create-view-content'

/** Main API service to get data for code insights */
export class InsightsAPI implements ApiService {
    /** Get combined (backend and extensions) code insights */
    public getCombinedViews = (
        getExtensionsInsights: () => Observable<ViewProviderResult[]>
    ): Observable<ViewInsightProviderResult[]> =>
        combineLatest([
            getExtensionsInsights().pipe(
                map(extensionInsights =>
                    extensionInsights.map(insight => ({ ...insight, source: ViewInsightProviderSourceType.Extension }))
                )
            ),
            fetchBackendInsights().pipe(
                map(backendInsights =>
                    backendInsights.map(
                        (insight, index): ViewInsightProviderResult => ({
                            id: `Backend insight ${index + 1}`,
                            view: {
                                title: insight.title,
                                subtitle: insight.description,
                                content: [createViewContent(insight)],
                            },
                            source: ViewInsightProviderSourceType.Backend,
                        })
                    )
                ),
                catchError(error =>
                    of<ViewInsightProviderResult[]>([
                        {
                            id: 'Backend insight',
                            view: asError(error),
                            source: ViewInsightProviderSourceType.Backend,
                        },
                    ])
                )
            ),
        ]).pipe(map(([extensionViews, backendInsights]) => [...backendInsights, ...extensionViews]))

    public getInsightCombinedViews = (
        extensionApi: Promise<Remote<FlatExtensionHostAPI>>
    ): Observable<ViewInsightProviderResult[]> =>
        this.getCombinedViews(() =>
            from(extensionApi).pipe(
                switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getInsightsViews({})))
            )
        )
}
