import { HTMLAttributes, FC } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useDeepMemo, BarChart, LegendList, LegendItem, useDebounce } from '@sourcegraph/wildcard'

import { GroupByField } from '../../../../../../../graphql-operations'
import {
    LivePreviewUpdateButton,
    LivePreviewCard,
    LivePreviewLoading,
    LivePreviewChart,
    LivePreviewBlurBackdrop,
    LivePreviewBanner,
    getSanitizedRepositories,
    COMPUTE_MOCK_CHART,
} from '../../../../../components'
import { CategoricalChartContent, SearchBasedInsightSeries } from '../../../../../core'
import { LivePreviewStatus, useLivePreviewComputeInsight } from '../../../../../core/hooks/live-preview-insight'

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    group?: string
}

interface ComputeLivePreviewProps extends HTMLAttributes<HTMLElement> {
    disabled: boolean
    repositories: string
    className?: string
    groupBy: GroupByField
    series: SearchBasedInsightSeries[]
}

export const ComputeLivePreview: FC<ComputeLivePreviewProps> = props => {
    const { disabled, repositories, series, groupBy, ...attribute } = props

    const settings = useDebounce(
        useDeepMemo({
            groupBy,
            repositories: getSanitizedRepositories(repositories),
            series: series.map(srs => ({
                query: srs.query,
                label: srs.name,
                stroke: srs.stroke ?? 'blue',
            })),
        }),
        500
    )

    const { state, refetch } = useLivePreviewComputeInsight({
        skip: disabled,
        ...settings,
    })

    return (
        <aside {...attribute}>
            <LivePreviewUpdateButton disabled={disabled} onClick={refetch} />

            <LivePreviewCard>
                {state.status === LivePreviewStatus.Loading ? (
                    <LivePreviewLoading>Loading code insight</LivePreviewLoading>
                ) : state.status === LivePreviewStatus.Error ? (
                    <ErrorAlert error={state.error} />
                ) : (
                    <LivePreviewChart>
                        {parent =>
                            state.status === LivePreviewStatus.Data ? (
                                <BarChart
                                    width={parent.width}
                                    height={parent.height}
                                    data={state.data}
                                    getCategory={(datum: LanguageUsageDatum) => datum.group}
                                    getDatumName={(datum: LanguageUsageDatum) => datum.name}
                                    getDatumValue={(datum: LanguageUsageDatum) => datum.value}
                                    getDatumColor={(datum: LanguageUsageDatum) => datum.fill}
                                />
                            ) : (
                                <>
                                    <LivePreviewBlurBackdrop
                                        as={BarChart}
                                        width={parent.width}
                                        height={parent.height}
                                        getCategory={(datum: unknown) => (datum as LanguageUsageDatum).group}
                                        // We cast to unknown here because ForwardReferenceComponent
                                        // doesn't support types inferring if component has a generic parameter.
                                        {...(COMPUTE_MOCK_CHART as CategoricalChartContent<unknown>)}
                                    />
                                    <LivePreviewBanner>You’ll see your insight’s chart preview here</LivePreviewBanner>
                                </>
                            )
                        }
                    </LivePreviewChart>
                )}

                {state.status === LivePreviewStatus.Data && (
                    <LegendList className="mt-3">
                        {settings.series.map(series => (
                            <LegendItem key={series.label} color={series.stroke} name={series.label} />
                        ))}
                    </LegendList>
                )}
            </LivePreviewCard>
        </aside>
    )
}
