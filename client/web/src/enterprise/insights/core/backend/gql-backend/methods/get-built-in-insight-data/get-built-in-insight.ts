import { from, Observable } from 'rxjs'

import { GetBuiltInsightInput, InsightContent } from '../../../code-insights-backend-types'

import { getLangStatsInsightContent } from './get-lang-stats-insight-content'

export function getBuiltInInsight(input: GetBuiltInsightInput): Observable<InsightContent<any>> {
    const { insight } = input

    return from(getLangStatsInsightContent(insight))
}
