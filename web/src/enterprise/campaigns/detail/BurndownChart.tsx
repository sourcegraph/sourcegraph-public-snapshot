import H from 'history'
import React from 'react'
import {
    Area,
    ComposedChart,
    LabelFormatter,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
    TooltipPayload,
} from 'recharts'
import { ICampaign } from '../../../../../shared/src/graphql/schema'

interface Props extends Pick<ICampaign, 'changesetCountsOverTime'> {
    history: H.History
}

const dateTickFormat = new Intl.DateTimeFormat(undefined, { month: 'long', day: 'numeric' })
const dateTickFormatter = (timestamp: number): string => dateTickFormat.format(timestamp)

// const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' })
const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
const tooltipLabelFormatter = (date: number): string => tooltipLabelFormat.format(date)

const toLocaleString = (value: number): string => value.toLocaleString()

const tooltipStyle: React.CSSProperties = {
    color: 'var(--body-color)',
    border: 'none',
    background: 'var(--body-bg)',
}

const commonAreaProps = {
    isAnimationActive: false,
    strokeWidth: 0,
    stackId: 'stack',
    type: 'stepBefore',
} as const

const tooltipItemOrder: Record<string, number> = {
    'Open & awaiting review': 4,
    'Open & changes requested': 3,
    'Open & approved': 2,
    Closed: 1,
    Merged: 0,
}

const tooltipItemSorter = (item: TooltipPayload): number => tooltipItemOrder[item.name]

/**
 * A burndown chart showing progress of the campaigns changesets.
 */
export const CampaignBurndownChart: React.FunctionComponent<Props> = ({ changesetCountsOverTime }) => {
    if (changesetCountsOverTime.length <= 1) {
        return (
            <p>
                <em>Burndown chart will be shown when there is more than 1 day of data.</em>
            </p>
        )
    }
    const hasEntries = changesetCountsOverTime.some(counts => counts.total > 0)
    if (!hasEntries) {
        return (
            <p>
                <em>Burndown chart will be shown when data is available.</em>
            </p>
        )
    }
    return (
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
                    name="Time"
                    tickFormatter={dateTickFormatter}
                    type="number"
                    stroke="var(--text-muted)"
                    scale="time"
                />
                <YAxis
                    tickFormatter={toLocaleString}
                    stroke="var(--text-muted)"
                    type="number"
                    allowDecimals={false}
                    domain={[0, 'dataMax']}
                />
                <Tooltip
                    labelFormatter={tooltipLabelFormatter as LabelFormatter}
                    isAnimationActive={false}
                    wrapperStyle={{ border: '1px solid var(--color-border)' }}
                    contentStyle={tooltipStyle}
                    labelStyle={{ fontWeight: 'bold' }}
                    itemStyle={tooltipStyle}
                    itemSorter={tooltipItemSorter}
                />

                <Area dataKey="openPending" name="Open & awaiting review" fill="var(--warning)" {...commonAreaProps} />
                <Area
                    dataKey="openChangesRequested"
                    name="Open & changes requested"
                    fill="var(--danger)"
                    {...commonAreaProps}
                />
                <Area dataKey="openApproved" name="Open & approved" fill="var(--success)" {...commonAreaProps} />
                <Area dataKey="closed" name="Closed" fill="var(--secondary)" {...commonAreaProps} />
                <Area dataKey="merged" name="Merged" fill="var(--merged)" {...commonAreaProps} />
            </ComposedChart>
        </ResponsiveContainer>
    )
}
