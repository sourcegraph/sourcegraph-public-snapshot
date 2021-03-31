import { combineLatest, Observable, of } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { requestGraphQL } from '../../backend/graphql'
import { InsightsResult, InsightFields } from '../../graphql-operations'
import { LineChartContent } from 'sourcegraph'
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/out/src/graphql/graphql'
import { asError } from '@sourcegraph/shared/out/src/util/errors'
import { ViewProviderResult } from '@sourcegraph/shared/out/src/api/extension/extensionHostApi'

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

export enum ViewInsightProviderSourceType {
    Backend = 'Backend',
    Extension = 'Extension',
}

export interface ViewInsightProviderResult extends ViewProviderResult {
    /** The source of view provider to distinguish between data from extension and data from backend */
    source: ViewInsightProviderSourceType
}

export function getCombinedViews(
    getExtensionsInsights: () => Observable<ViewProviderResult[]>
): Observable<ViewInsightProviderResult[]> {
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
                            content: [backendInsightToViewContent(insight)],
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
