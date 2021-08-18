import { Observable, of, throwError } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { fetchBackendInsights } from '../requests/fetch-backend-insights'
import { BackendInsightData, BackendInsightInputs } from '../types'
import { createViewContent } from '../utils/create-view-content'

export class InsightStillProcessingError extends Error {
    constructor(message: string = 'Your insight is being processed') {
        super(message)
        this.name = 'InProcessError'
    }
}

export function getBackendInsightById(props: BackendInsightInputs): Observable<BackendInsightData> {
    const { id, filters, series } = props

    return fetchBackendInsights([id], filters).pipe(
        switchMap(backendInsights => {
            if (backendInsights.length === 0) {
                return throwError(new InsightStillProcessingError())
            }

            return of(backendInsights[0])
        }),
        map(insight => ({
            id: insight.id,
            view: {
                title: insight.title,
                subtitle: insight.description,
                content: [createViewContent(insight, series)],
                isFetchingHistoricalData: insight.series.some(
                    line => line.status.pendingJobs > 0 || line.status.failedJobs > 0
                ),
            },
        }))
    )
}
