import { Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'

import { ViewContexts, ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { asError } from '@sourcegraph/shared/src/util/errors'

import { isSearchBasedInsight } from '../../types'
import { GetBuiltInsightInput } from '../code-insights-backend-types'

import { getLangStatsInsightContent } from './get-lang-stats-insight-content'
import { getSearchInsightContent } from './get-search-insight-content/get-search-insight-content'

export function getBuiltInInsight<D extends keyof ViewContexts>(
    input: GetBuiltInsightInput<D>
): Observable<ViewProviderResult> {
    const { insight, options } = input
    return of(insight).pipe(
        // TODO Implement declarative fetchers map by insight type
        switchMap(insight =>
            isSearchBasedInsight(insight)
                ? getSearchInsightContent(insight, options)
                : getLangStatsInsightContent(insight, options)
        ),
        map(data => ({
            id: insight.id,
            view: {
                title: insight.title,
                content: [data],
            },
        })),
        catchError(error =>
            of<ViewProviderResult>({
                id: insight.id,
                view: asError(error),
            })
        )
    )
}
