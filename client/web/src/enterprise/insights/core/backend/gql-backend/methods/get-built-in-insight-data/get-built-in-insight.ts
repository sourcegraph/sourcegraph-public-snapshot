import { Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'

import { asError } from '@sourcegraph/common'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { isSearchBasedInsight } from '../../../../types'
import { GetBuiltInsightInput } from '../../../code-insights-backend-types'

import { getLangStatsInsightContent } from './get-lang-stats-insight-content'
import { getSearchInsightContent } from './get-search-insight-content'

export function getBuiltInInsight(input: GetBuiltInsightInput): Observable<ViewProviderResult> {
    const { insight } = input

    return of(insight).pipe(
        // TODO Implement declarative fetchers map by insight type
        switchMap(insight =>
            isSearchBasedInsight(insight) ? getSearchInsightContent(insight) : getLangStatsInsightContent(insight)
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
