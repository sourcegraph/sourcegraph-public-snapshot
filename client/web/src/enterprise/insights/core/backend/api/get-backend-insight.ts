import { Observable, of, throwError } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { SearchBackendBasedInsight } from '../../types/insight/search-insight'
import { fetchBackendInsights } from '../requests/fetch-backend-insights'
import { BackendInsightData } from '../types'
import { createLineChartContent } from '../utils/create-line-chart-content'

export class InsightStillProcessingError extends Error {
    constructor(message: string = 'Your insight is being processed') {
        super(message)
        this.name = 'InProcessError'
    }
}

export function getBackendInsight(insight: SearchBackendBasedInsight): Observable<BackendInsightData> {
    const { id, filters, series } = insight

    return fetchBackendInsights([id], filters).pipe(
        switchMap(backendInsights => {
            if (backendInsights.length === 0) {
                return throwError(new InsightStillProcessingError())
            }

            return of(backendInsights[0])
        }),
        map(backendInsight => ({
            id: backendInsight.id,
            view: {
                title: insight.title ?? backendInsight.title,
                subtitle: '',
                content: [createLineChartContent(backendInsight, series)],
                isFetchingHistoricalData: backendInsight.series.some(
                    ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
                ),
            },
        }))
    )
}
