import { Observable, of } from 'rxjs'
import { catchError, map, startWith } from 'rxjs/operators'

import { asError } from '@sourcegraph/shared/src/util/errors'

import { fetchBackendInsights } from '../requests/fetch-backend-insights'
import { ViewInsightProviderResult, ViewInsightProviderSourceType } from '../types'
import { createViewContent } from '../utils/create-view-content'

/**
 * Returns list of backend insights via gql API request.
 */
export function getBackendInsights(insightIds?: string[]): Observable<ViewInsightProviderResult[]> {
    // If Ids field wasn't specified then return all insights
    if (!insightIds) {
        return getRawBackendInsights([]).pipe(
            startWith([
                {
                    id: 'Backend insights',
                    view: undefined,
                    source: ViewInsightProviderSourceType.Backend,
                },
            ])
        )
    }

    if (insightIds.length === 0) {
        return of([])
    }

    return getRawBackendInsights(insightIds).pipe(
        startWith(
            insightIds.map(id => ({
                id,
                view: undefined,
                source: ViewInsightProviderSourceType.Backend,
            }))
        )
    )
}

function getRawBackendInsights(insightIds: string[]): Observable<ViewInsightProviderResult[]> {
    return fetchBackendInsights(insightIds).pipe(
        map(backendInsights =>
            backendInsights.map(
                (insight): ViewInsightProviderResult => ({
                    id: insight.id,
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
            of<ViewInsightProviderResult[]>(
                insightIds.map(id => ({
                    id,
                    view: asError(error),
                    source: ViewInsightProviderSourceType.Backend,
                }))
            )
        )
    )
}
