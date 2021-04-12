import { combineLatest, from, Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { asError } from '@sourcegraph/shared/out/src/util/errors'
import { ViewProviderResult } from '@sourcegraph/shared/out/src/api/extension/extensionHostApi'
import { createViewContent } from './utils/create-view-content';
import { fetchBackendInsights } from './requests/fetch-backend-insights'
import { ApiService, ViewInsightProviderResult, ViewInsightProviderSourceType } from './types';
import { Remote } from 'comlink';
import { FlatExtensionHostAPI } from '@sourcegraph/shared/out/src/api/contract';
import { wrapRemoteObservable } from '@sourcegraph/shared/out/src/api/client/api/common';

/** Main API service to get data for code insights */
export class InsightsAPI implements ApiService {

    /** Get combined (backend and extensions) code insights */
    public getCombinedViews = (getExtensionsInsights: () => Observable<ViewProviderResult[]>): Observable<ViewInsightProviderResult[]> => {
        return combineLatest([
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
    }

    public getInsightCombinedViews = (extensionApi: Promise<Remote<FlatExtensionHostAPI>>) => {
        return this.getCombinedViews(
            () =>
                from(extensionApi).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getInsightsViews({})))
                )
        )
    }
}

