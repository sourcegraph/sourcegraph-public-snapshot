import { Observable, of, throwError } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { LineChartContent } from 'sourcegraph'

import { fetchBackendInsights } from '../../../../../core/backend/requests/fetch-backend-insights'
import { createViewContent } from '../../../../../core/backend/utils/create-view-content'

export type BackendInsightContent = LineChartContent<{ dateTime: number; [seriesKey: string]: number }, 'dateTime'>

export interface BackendInsightData {
    id: string
    view: {
        title: string
        subtitle: string
        content: BackendInsightContent[]
    }
}
/**
 * Returns backend insight data via gql API request.
 */
export function getBackendInsightById(id: string): Observable<BackendInsightData> {
    return fetchBackendInsights([id]).pipe(
        switchMap(backendInsights => {
            if (backendInsights.length === 0) {
                return throwError(new Error("We couldn't find insight"))
            }

            return of(backendInsights[0])
        }),
        map(
            (insight): BackendInsightData => ({
                id: insight.id,
                view: {
                    title: insight.title,
                    subtitle: insight.description,
                    content: [createViewContent(insight)],
                },
            })
        )
    )
}
