import { combineLatest, Observable, of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { requestGraphQL } from '../backend/graphql'
import { InsightsResult, InsightFields } from '../graphql-operations'
import { LineChartContent } from 'sourcegraph'
import {
    getViewsForContainer,
    ViewContexts,
    ViewProviderResult,
    ViewService,
} from '../../../shared/src/api/client/services/viewService'
import { ContributableViewContainer } from '../../../shared/src/api/protocol'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import { asError } from '../../../shared/src/util/errors'

const insightFieldsFragment = gql`
    fragment InsightFields on Insight {
        title
        description
        series {
            label
            points {
                dateTime
                value
            }
        }
    }
`
function fetchBackendInsights(): Observable<InsightFields[]> {
    return requestGraphQL<InsightsResult>(gql`
        query Insights {
            insights {
                nodes {
                    ...InsightFields
                }
            }
        }
        ${insightFieldsFragment}
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.insights?.nodes ?? [])
    )
}

export function getCombinedViews<W extends ContributableViewContainer>(
    where: W,
    parameters: ViewContexts[W],
    viewService: Pick<ViewService, 'getWhere'>
): Observable<ViewProviderResult[]> {
    return combineLatest([
        getViewsForContainer(where, parameters, viewService),
        fetchBackendInsights().pipe(
            map(backendInsights =>
                backendInsights.map(
                    (insight, index): ViewProviderResult => ({
                        id: `insight.backend${index}`,
                        view: {
                            title: insight.title,
                            subtitle: insight.description,
                            content: [backendInsightToViewContent(insight)],
                        },
                    })
                )
            ),
            catchError(error => of<ViewProviderResult[]>([{ id: 'insight.backend', view: asError(error) }]))
        ),
    ]).pipe(map(([extensionViews, backendInsights]) => [...backendInsights, ...extensionViews]))
}

function backendInsightToViewContent(
    insight: InsightFields
): LineChartContent<{ dateTime: number; [seriesKey: string]: number }, 'dateTime'> {
    const dataByXValue = new Map<string, { dateTime: number; [seriesKey: string]: number }>()
    for (const [seriesIndex, series] of insight.series.entries()) {
        for (const point of series.points) {
            let dataObject = dataByXValue.get(point.dateTime)
            if (!dataObject) {
                dataObject = {
                    dateTime: Date.parse(point.dateTime),
                }
                dataByXValue.set(point.dateTime, dataObject)
            }
            dataObject[`series${seriesIndex}`] = point.value
        }
    }
    return {
        chart: 'line',
        data: [...dataByXValue.values()],
        series: insight.series.map((series, index) => ({
            name: series.label,
            dataKey: `series${index}`,
        })),
        xAxis: {
            dataKey: 'dateTime',
            scale: 'time',
            type: 'number',
        },
    }
}
