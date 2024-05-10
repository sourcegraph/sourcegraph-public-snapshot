import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'

import { AnalyticsDateRange, AnalyticsGrouping } from '../../graphql-operations'
import { useURLSyncedState } from '../../hooks'

type IAggregation = 'count' | 'uniqueUsers'

type ChartName =
    | 'BatchChanges'
    | 'Insights'
    | 'CodeIntel'
    | 'Extensions'
    | 'Notebooks'
    | 'Overview'
    | 'Search'
    | 'Users'

const v2ChartTypes: { [key in ChartName]: number } = {
    BatchChanges: 1,
    Insights: 2,
    CodeIntel: 3,
    Extensions: 4,
    Notebooks: 5,
    Overview: 6,
    Search: 7,
    Users: 8,
}
const v2DateRangeTypes: { [key in AnalyticsDateRange]: number } = {
    CUSTOM: 1,
    LAST_MONTH: 2,
    LAST_THREE_MONTHS: 3,
    LAST_WEEK: 4,
}
const v2AggregationTypes: { [key in IAggregation]: number } = {
    count: 1,
    uniqueUsers: 2,
}
const v2GroupingTypes: { [key in AnalyticsGrouping]: number } = {
    DAILY: 1,
    WEEKLY: 2,
}

interface IProps extends TelemetryV2Props {
    name: ChartName
    aggregation?: IAggregation
    dateRange?: AnalyticsDateRange
    grouping?: AnalyticsGrouping
}

interface IResult {
    dateRange: {
        value: AnalyticsDateRange
        label: string
        onChange: (value: AnalyticsDateRange) => void
        items: { value: AnalyticsDateRange; label: string }[]
    }
    aggregation: {
        selected: IAggregation
        onChange: (value: IAggregation) => void
        items: { value: IAggregation; label: string; tooltip: string }[]
    }
    grouping: {
        value: AnalyticsGrouping
        label: string
        onChange: (value: AnalyticsGrouping) => void
        items: { value: AnalyticsGrouping; label: string }[]
    }
}

export function useChartFilters(props: IProps): IResult {
    const [data, setData] = useURLSyncedState({
        dateRange: props.dateRange || AnalyticsDateRange.LAST_THREE_MONTHS,
        aggregation: props.aggregation || 'count',
        grouping: props.grouping || AnalyticsGrouping.WEEKLY,
    })

    return {
        dateRange: {
            value: data.dateRange,
            label: 'Date range',
            onChange: value => {
                setData({
                    dateRange: value,
                    grouping:
                        value === AnalyticsDateRange.LAST_WEEK ? AnalyticsGrouping.DAILY : AnalyticsGrouping.WEEKLY,
                })
                EVENT_LOGGER.log(`AdminAnalytics${props.name}DateRange${value}`)
                props.telemetryRecorder.recordEvent('admin.analytics.dateRange', 'change', {
                    metadata: { kind: v2ChartTypes[props.name], value: v2DateRangeTypes[value] },
                })
            },
            items: [
                { value: AnalyticsDateRange.LAST_WEEK, label: 'Last week' },
                { value: AnalyticsDateRange.LAST_MONTH, label: 'Last month' },
                { value: AnalyticsDateRange.LAST_THREE_MONTHS, label: 'Last 3 months' },
            ],
        },
        aggregation: {
            selected: data.aggregation,
            onChange: value => {
                setData({ aggregation: value })
                EVENT_LOGGER.log(`AdminAnalytics${props.name}Aggregate${value === 'count' ? 'Totals' : 'Uniques'}`)
                props.telemetryRecorder.recordEvent('admin.analytics.aggregation', 'change', {
                    metadata: { kind: v2ChartTypes[props.name], value: v2AggregationTypes[value] },
                })
            },
            items: [
                {
                    tooltip: 'total # of actions triggered',
                    label: 'Total events',
                    value: 'count',
                },
                {
                    tooltip: 'unique # of users triggered',
                    label: 'Unique users',
                    value: 'uniqueUsers',
                },
            ],
        },
        grouping: {
            value: data.grouping,
            label: 'Display as',
            onChange: value => {
                setData({ grouping: value })
                EVENT_LOGGER.log(`AdminAnalytics${props.name}DisplayAs${value}`)
                props.telemetryRecorder.recordEvent('admin.analytics.grouping', 'change', {
                    metadata: { kind: v2ChartTypes[props.name], value: v2GroupingTypes[value] },
                })
            },
            items: [
                { value: AnalyticsGrouping.DAILY, label: 'Daily' },
                { value: AnalyticsGrouping.WEEKLY, label: 'Weekly' },
            ],
        },
    }
}
