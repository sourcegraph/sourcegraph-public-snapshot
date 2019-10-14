import H from 'history'
import React from 'react'
import { Area, ComposedChart, LabelFormatter, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { ICampaign } from '../../../../../shared/src/graphql/schema'

interface Props extends Pick<ICampaign, 'changesetCountsOverTime'> {
    history: H.History
}

const dateTickFormat = new Intl.DateTimeFormat(undefined, { month: 'long', day: 'numeric' })
const dateTickFormatter = (timestamp: number): string => dateTickFormat.format(timestamp)

// const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' })
const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
const tooltipLabelFormatter = (date: number): string => tooltipLabelFormat.format(date)

const STYLE: React.CSSProperties = {
    color: 'var(--body-color)',
    backgroundColor: 'var(--body-bg)',
}

const toLocaleString = (value: number): string => value.toLocaleString()

/**
 * A burndown chart showing progress of the campaigns changesets.
 */
export const CampaignBurndownChart: React.FunctionComponent<Props> = ({ changesetCountsOverTime }) =>
    changesetCountsOverTime.length <= 1 ? (
        <div className="alert alert-info">Burndown chart will be shown when there is more than 1 day of data.</div>
    ) : (
        <ResponsiveContainer width="100%" height={300}>
            <ComposedChart
                data={changesetCountsOverTime.map(snapshot => ({ ...snapshot, date: Date.parse(snapshot.date) }))}
            >
                <XAxis
                    dataKey="date"
                    domain={[
                        changesetCountsOverTime[0].date,
                        changesetCountsOverTime[changesetCountsOverTime.length - 1].date,
                    ]}
                    // TODO!(sqs): delete? domain={[startDate, startDate + openThreads.length * 24 * 60 * 60 * 1000]}
                    name="Time"
                    tickFormatter={dateTickFormatter}
                    type="number"
                    stroke="var(--text-muted)"
                />
                <YAxis
                    tickFormatter={toLocaleString}
                    stroke="var(--text-muted)"
                    type="number"
                    allowDecimals={false}
                    domain={[0, 'dataMax']}
                />
                <Tooltip
                    // formatter={tooltipFormatter}
                    labelFormatter={tooltipLabelFormatter as LabelFormatter}
                    isAnimationActive={false}
                    wrapperStyle={STYLE}
                    itemStyle={STYLE}
                    labelStyle={STYLE}
                />

                <Area
                    type="step"
                    dataKey="openApproved"
                    name="Open & approved"
                    fill="var(--success)"
                    strokeWidth={0}
                    isAnimationActive={false}
                />
                <Area
                    stackId="threadState"
                    type="step"
                    dataKey="mergedThreads"
                    name="Merged"
                    fill="var(--merged)"
                    strokeWidth={0}
                    isAnimationActive={false}
                />
                <Area
                    stackId="threadState"
                    type="step"
                    dataKey="closedThreads"
                    name="Closed"
                    fill="var(--secondary)"
                    strokeWidth={0}
                    isAnimationActive={false}
                />
            </ComposedChart>
        </ResponsiveContainer>
    )
