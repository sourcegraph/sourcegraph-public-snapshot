import { format } from 'date-fns'
import H from 'history'
import React from 'react'
import {
    Area,
    ComposedChart,
    LabelFormatter,
    ResponsiveContainer,
    TickFormatterFunction,
    Tooltip,
    XAxis,
    YAxis,
} from 'recharts'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { numberWithCommas } from '../../../../../../shared/src/util/strings'
import { DismissibleAlert } from '../../../../components/DismissibleAlert'
import { useCampaignBurndownChart } from './useCampaignBurndownChart'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
    history: H.History
}

const dateTickFormatter: TickFormatterFunction = date => format(date, 'MMM d')

const tooltipLabelFormatter: LabelFormatter = date => format(date as number, 'PP')

const STYLE: React.CSSProperties = {
    color: 'var(--body-color)',
    backgroundColor: 'var(--body-bg)',
}

const LOADING = 'loading' as const

/**
 * A burndown chart showing progress toward closing a campaign's threads.
 */
export const CampaignBurndownChart: React.FunctionComponent<Props> = ({ campaign, className = '' }) => {
    const [burndownChart] = useCampaignBurndownChart(campaign)
    return (
        <div className={`campaign-burndown-chart ${className}`}>
            {isErrorLike(burndownChart) ? (
                <div className="alert alert-danger">Error generating burndown chart: {burndownChart.message}</div>
            ) : burndownChart !== LOADING && burndownChart.dates.length <= 1 ? (
                <DismissibleAlert partialStorageKey="CampaignBurndownChart.insufficientData" className="alert-info">
                    Burndown chart will be shown when there is more than 1 day of data.
                </DismissibleAlert>
            ) : (
                <ResponsiveContainer width="100%" height={300}>
                    <ComposedChart
                        data={
                            burndownChart !== LOADING && !isErrorLike(burndownChart)
                                ? burndownChart.dates.map((date, i) => ({
                                      date: Date.parse(date),
                                      openThreads: burndownChart.openThreads[i],
                                      mergedThreads: burndownChart.mergedThreads[i],
                                      closedThreads: burndownChart.closedThreads[i],
                                      openApprovedThreads: burndownChart.openApprovedThreads[i],
                                  }))
                                : []
                        }
                    >
                        <XAxis
                            dataKey="date"
                            domain={
                                burndownChart !== LOADING &&
                                !isErrorLike(burndownChart) &&
                                burndownChart.dates.length > 0
                                    ? [burndownChart.dates[0], burndownChart.dates[burndownChart.dates.length - 1]]
                                    : [0, 0]
                            }
                            // TODO!(sqs): delete? domain={[startDate, startDate + openThreads.length * 24 * 60 * 60 * 1000]}
                            name="Time"
                            tickFormatter={dateTickFormatter}
                            type="number"
                            stroke="var(--text-muted)"
                        />
                        <YAxis
                            tickFormatter={numberWithCommas}
                            stroke="var(--text-muted)"
                            type="number"
                            allowDecimals={false}
                            domain={[0, 'dataMax']}
                        />
                        <Tooltip
                            // formatter={tooltipFormatter}
                            labelFormatter={tooltipLabelFormatter}
                            isAnimationActive={false}
                            wrapperStyle={STYLE}
                            itemStyle={STYLE}
                            labelStyle={STYLE}
                        />
                        <Area
                            type="step"
                            dataKey="openApprovedThreads"
                            name="Open & approved"
                            fill="var(--info)"
                            strokeWidth={0}
                            isAnimationActive={false}
                        />
                        <Area
                            stackId="threadState"
                            type="step"
                            dataKey="openThreads"
                            name="Open changesets"
                            stroke="var(--body-color)"
                            strokeWidth={3}
                            fill="transparent"
                            activeDot={{ r: 5 }}
                            isAnimationActive={false}
                        />
                        <Area
                            stackId="threadState"
                            type="step"
                            dataKey="mergedThreads"
                            name="Merged"
                            fill="var(--success)"
                            strokeWidth={0}
                            isAnimationActive={false}
                        />
                        <Area
                            stackId="threadState"
                            type="step"
                            dataKey="closedThreads"
                            name="Closed"
                            fill="var(--text-muted)"
                            strokeWidth={0}
                            isAnimationActive={false}
                        />
                    </ComposedChart>
                </ResponsiveContainer>
            )}
        </div>
    )
}
