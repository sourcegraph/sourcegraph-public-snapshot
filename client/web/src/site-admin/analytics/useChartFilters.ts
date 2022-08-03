import { useState } from 'react'

import { AnalyticsDateRange, AnalyticsGrouping } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

type IAggregation = 'count' | 'uniqueUsers'

interface IProps {
    name: string
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
    const [aggregation, setAggregation] = useState<IAggregation>(props.aggregation || 'count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(
        props.dateRange || AnalyticsDateRange.LAST_THREE_MONTHS
    )
    const [grouping, setGrouping] = useState<AnalyticsGrouping>(props.grouping || AnalyticsGrouping.WEEKLY)

    return {
        dateRange: {
            value: dateRange,
            label: 'Date range',
            onChange: value => {
                setDateRange(value)
                if (value === AnalyticsDateRange.LAST_WEEK) {
                    setGrouping(AnalyticsGrouping.DAILY)
                } else {
                    setGrouping(AnalyticsGrouping.WEEKLY)
                }
                eventLogger.log(`AdminAnalytics${props.name}DateRange${value}`)
            },
            items: [
                { value: AnalyticsDateRange.LAST_WEEK, label: 'Last week' },
                { value: AnalyticsDateRange.LAST_MONTH, label: 'Last month' },
                { value: AnalyticsDateRange.LAST_THREE_MONTHS, label: 'Last 3 months' },
            ],
        },
        aggregation: {
            selected: aggregation,
            onChange: value => {
                setAggregation(value)
                eventLogger.log(`AdminAnalytics${props.name}Aggregate${value === 'count' ? 'Totals' : 'Uniques'}`)
            },
            items: [
                {
                    tooltip: 'total # of actions triggered',
                    label: 'Total events',
                    value: 'count',
                },
                {
                    tooltip: 'unique # of users triggered',
                    label: 'Uniques users',
                    value: 'uniqueUsers',
                },
            ],
        },
        grouping: {
            value: grouping,
            label: 'Display as',
            onChange: value => {
                setGrouping(value)
                eventLogger.log(`AdminAnalytics${props.name}DisplayAs${value}`)
            },
            items: [
                { value: AnalyticsGrouping.DAILY, label: 'Daily' },
                { value: AnalyticsGrouping.WEEKLY, label: 'Weekly' },
            ],
        },
    }
}
