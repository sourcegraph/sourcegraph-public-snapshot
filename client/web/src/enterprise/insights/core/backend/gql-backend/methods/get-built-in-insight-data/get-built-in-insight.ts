import { Observable, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { isSearchBasedInsight } from '../../../../types'
import { GetBuiltInsightInput, InsightContent } from '../../../code-insights-backend-types'

import { getLangStatsInsightContent } from './get-lang-stats-insight-content'
import { getSearchInsightContent } from './get-search-insight-content'

export function getBuiltInInsight(input: GetBuiltInsightInput): Observable<InsightContent<any>> {
    const { insight } = input

    return of(insight).pipe(
        // TODO Implement declarative fetchers map by insight type
        switchMap(insight =>
            isSearchBasedInsight(insight) ? getSearchInsightContent(insight) : getLangStatsInsightContent(insight)
        )
    )
}
