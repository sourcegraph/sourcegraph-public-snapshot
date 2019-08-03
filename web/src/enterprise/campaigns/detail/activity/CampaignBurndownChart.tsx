import { format } from 'date-fns'
import H from 'history'
import React from 'react'
import {
    Area,
    AreaChart,
    CartesianGrid,
    ComposedChart,
    LabelFormatter,
    Legend,
    Line,
    LineChart,
    ReferenceLine,
    ResponsiveContainer,
    TickFormatterFunction,
    Tooltip,
    TooltipFormatter,
    XAxis,
    YAxis,
} from 'recharts'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { numberWithCommas, pluralize } from '../../../../../../shared/src/util/strings'

interface Props extends ExtensionsControllerProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
    history: H.History
}

const openThreads = [1071, 1218, 1231, 1121, 1018, 980, 979, 930, 715, 331, 371, 102, 81, 81, 81, 31, 23, 7, 7, 3]

const approvedThreads = openThreads.map((n, i) => Math.floor((i / openThreads.length) * n))

const ciFailingThreads = approvedThreads.map((n, i) => Math.floor((openThreads[i] - n) * 0.87))

const errorThreads = approvedThreads.map((n, i) => Math.floor((openThreads[i] - n) * 0.13))

const closedThreads = openThreads.map((n, i) =>
    Math.floor(Math.max(...openThreads.slice(0, i + 1)) - n + Math.pow(2, 1 + i / 3))
)

const startDate = Date.now() - openThreads.length * 24 * 60 * 60 * 1000

const data: {
    date: number
    openThreads: number
    approvedThreads: number
    ciFailingThreads: number
    errorThreads: number
}[] = openThreads.map((openThreads, i) => ({
    date: startDate + i * 24 * 60 * 60 * 1000,
    openThreads,
    approvedThreads: approvedThreads[i],
    ciFailingThreads: ciFailingThreads[i],
    errorThreads: errorThreads[i],
    closedThreads: closedThreads[i],
}))

const dateTickFormatter: TickFormatterFunction = date => format(date, 'MMM d')

const tooltipLabelFormatter: LabelFormatter = (date: number) => format(date, 'PP')

const tooltipFormatter: TooltipFormatter = (value: number) => [
    `${numberWithCommas(value)} open ${pluralize('changeset', value)}`,
]

const STYLE: React.CSSProperties = {
    color: 'var(--body-color)',
    backgroundColor: 'var(--body-bg)',
}

/**
 * A burndown chart showing progress toward closing a campaign's threads.
 */
export const CampaignBurndownChart: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => (
    <div className={`campaign-burndown-chart ${className}`}>
        <ResponsiveContainer width="100%" height={300}>
            <ComposedChart data={data}>
                <XAxis
                    dataKey="date"
                    // domain={['auto', 'auto']}
                    domain={[startDate, startDate + openThreads.length * 24 * 60 * 60 * 1000]}
                    name="Time"
                    tickFormatter={dateTickFormatter}
                    type="number"
                    stroke="var(--text-muted)"
                />
                <YAxis tickFormatter={numberWithCommas} stroke="var(--text-muted)" type="number" />
                <Tooltip
                    // formatter={tooltipFormatter}
                    labelFormatter={tooltipLabelFormatter}
                    isAnimationActive={false}
                    wrapperStyle={STYLE}
                    itemStyle={STYLE}
                    labelStyle={STYLE}
                />
                <Area
                    stackId="openThreads"
                    type="step"
                    dataKey="approvedThreads"
                    name="Approved"
                    fill="var(--success)"
                    strokeWidth={0}
                />
                <Area
                    stackId="openThreads"
                    type="step"
                    dataKey="ciFailingThreads"
                    name="Failing CI"
                    fill="var(--danger)"
                    strokeWidth={0}
                />
                <Area
                    stackId="openThreads"
                    type="step"
                    dataKey="errorThreads"
                    name="Error"
                    fill="var(--warning)"
                    strokeWidth={0}
                />
                <Area
                    stackId="openThreads"
                    type="step"
                    dataKey="closedThreads"
                    name="Closed"
                    fill="var(--text-muted)"
                    strokeWidth={0}
                />
                <Line
                    type="step"
                    dataKey="openThreads"
                    name="Open changesets"
                    stroke="var(--body-color)"
                    strokeWidth={4}
                    strokeOpacity={1}
                    activeDot={{ r: 5 }}
                    isAnimationActive={false}
                />
                <ReferenceLine
                    y={openThreads[0]}
                    strokeWidth={2}
                    strokeOpacity={0.7}
                    fontWeight="bold"
                    style={STYLE}
                    color="var(--info)"
                    stroke="var(--info)"
                    strokeDasharray="10 2"
                />
            </ComposedChart>
        </ResponsiveContainer>
    </div>
)
