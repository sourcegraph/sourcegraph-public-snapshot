/* eslint-disable react/forbid-dom-props */
import { useMemo, useState } from 'react'

import { mdiChartLineVariant, mdiChartTimelineVariantShimmer } from '@mdi/js'
import classNames from 'classnames'
import { addDays, getDayOfYear, startOfDay, startOfWeek, sub } from 'date-fns'
import { upperFirst } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useQuery } from '@sourcegraph/http-client'
import {
    H2,
    Card,
    Select,
    Input,
    H3,
    Text,
    Icon,
    ButtonGroup,
    Button,
    LoadingSpinner,
    Tooltip,
} from '@sourcegraph/wildcard'

import { LineChart, ParentSize, Series } from '../../charts'
import { BarChart } from '../../charts/components/bar-chart/BarChart'
import {
    AnalyticsDateRange,
    SearchStatisticsResult,
    SearchStatisticsVariables,
    NotebooksStatisticsResult,
    NotebooksStatisticsVariables,
    UsersStatisticsResult,
    UsersStatisticsVariables,
} from '../../graphql-operations'

import { formatNumber } from './format-number'
import { SEARCH_STATISTICS, NOTEBOOKS_STATISTICS, USERS_STATISTICS } from './queries'

import styles from './index.module.scss'

interface TimeSavedCalculatorProps {
    color: string
    label: string
    value: number
    minPerItem: number
    description?: string
}

const TimeSavedCalculator: React.FunctionComponent<TimeSavedCalculatorProps> = ({
    color,
    value,
    description,
    minPerItem: minsPerCount,
    label,
}) => {
    const [minutesPerCount, setMinutesPerCount] = useState(minsPerCount)
    const hoursSaved = useMemo(() => (value * minutesPerCount) / 60, [value, minutesPerCount])
    return (
        <Card className="mb-3 p-4 d-flex justify-content-between flex-row" key={label}>
            <div className={styles.calculatorInnerLeft}>
                <Text as="span" style={{ color }} className={classNames(styles.count, 'text-center')}>
                    {formatNumber(value)}
                </Text>
                <Input
                    type="number"
                    value={minutesPerCount}
                    className={styles.calculatorInput}
                    onChange={event => setMinutesPerCount(Number(event.target.value))}
                />
                <Text as="span" className={styles.count}>
                    {formatNumber(hoursSaved)}
                </Text>
                <Text as="span">{label}</Text>
                <Text as="span">
                    minutes saved
                    <br />
                    per action
                </Text>
                <Text as="span">hours saved</Text>
            </div>
            <div className="m-0 flex-1 d-flex flex-column justify-content-between">
                <Text as="span" className="font-weight-bold">
                    About this statistics
                </Text>
                <Text as="span">{description}</Text>
            </div>
        </Card>
    )
}

interface ChartContainerProps {
    className?: string
    title?: string
    labelX?: string
    labelY?: string
    children: (width: number) => React.ReactNode
}

const ChartContainer: React.FunctionComponent<ChartContainerProps> = ({
    className,
    title,
    children,
    labelX,
    labelY,
}) => (
    <div className={className}>
        {title && <Text className='text-center'>{title}</Text>}
        <div className="d-flex">
            {labelY && <span className={styles.chartYLabel}>{labelY}</span>}
            <ParentSize>{({ width }) => children(width)}</ParentSize>
        </div>
        {labelX && <div className={styles.chartXLabel}>{labelX}</div>}
    </div>
)

interface FrequencyDatum {
    label: string
    value: number
}

interface StandardDatum {
    date: Date
    value: number
}

interface ValueLegendItemProps {
    color: string
    description: string
    value: number
}

const ValueLegendItem: React.FunctionComponent<ValueLegendItemProps> = ({ value, color, description }) => (
    <div className="d-flex flex-column align-items-center mr-3 justify-content-center">
        <span style={{ color }} className={styles.count}>
            {formatNumber(value)}
        </span>
        <span className={classNames('text-center', styles.textWrap)}>{description}</span>
    </div>
)

interface ValueLegendListProps {
    className?: string
    items: (ValueLegendItemProps & { position?: 'left' | 'right' })[]
}

const ValueLegendList: React.FunctionComponent<ValueLegendListProps> = ({ items, className }) => (
    <div className={classNames('d-flex justify-content-between', className)}>
        <div className="d-flex justify-content-left">
            {items
                .filter(item => item.position !== 'right')
                .map(item => (
                    <ValueLegendItem key={item.description} {...item} />
                ))}
        </div>
        <div className="d-flex justify-content-right">
            {items
                .filter(item => item.position === 'right')
                .map(item => (
                    <ValueLegendItem key={item.description} {...item} />
                ))}
        </div>
    </div>
)

interface HorizontalSelect<T> {
    onChange: (value: T) => void
    value: T
    label: string
    className?: string
    items: { label: string; value: T; disabled?: boolean }[]
}

const HorizontalSelect = <T extends string>({
    items,
    label,
    value,
    onChange,
    className,
}: React.PropsWithChildren<HorizontalSelect<T>>): JSX.Element => (
    <Select
        id="date-range"
        label={label}
        isCustomStyle={true}
        className={classNames('d-flex align-items-center m-0', className)}
        labelClassName="mb-0"
        selectClassName="ml-2"
        value={value}
        onChange={value => onChange(value.target.value as T)}
    >
        {items.map(({ value, label, disabled }) => (
            <option key={label} value={value} disabled={disabled}>
                {label}
            </option>
        ))}
    </Select>
)

interface ToggleGroupProps<T> {
    selected: T
    className?: string
    items: {
        tooltip: string
        label: string
        value: T
    }[]
    onChange: (value: T) => void
}

const ToggleSelect = <T extends any>({
    selected,
    items,
    onChange,
    className,
}: React.PropsWithChildren<ToggleGroupProps<T>>): JSX.Element => (
    <ButtonGroup className={className}>
        {items.map(({ tooltip, label, value }) => (
            <Tooltip key={label} content={tooltip} placement="top">
                <Button
                    onClick={() => onChange(value)}
                    outline={selected !== value}
                    variant={selected !== value ? 'secondary' : 'primary'}
                    display="inline"
                    size="sm"
                >
                    {label}
                </Button>
            </Tooltip>
        ))}
    </ButtonGroup>
)

const AnalyticsPageTitle: React.FunctionComponent = ({ children }) => (
    <H2 className="mb-4 d-flex align-items-center">
        <Icon
            className="mr-1"
            color="var(--link-color)"
            svgPath={mdiChartLineVariant}
            size="sm"
            aria-label="Search Statistics"
        />
        {children}
    </H2>
)

export const AnalyticsComingSoon: React.FunctionComponent<RouteComponentProps<{}>> = props => {
    const title = useMemo(() => {
        const title = props.match.path.split('/').filter(Boolean)[2] ?? 'Overview'
        return upperFirst(title.replace('-', ' '))
    }, [props.match.path])
    return (
        <>
            <AnalyticsPageTitle>Analytics / {title}</AnalyticsPageTitle>

            <div className="d-flex flex-column justify-content-center align-items-center p-5">
                <Icon
                    svgPath={mdiChartTimelineVariantShimmer}
                    aria-label="Home analytics icon"
                    className={classNames(styles.largeIcon, 'm-3')}
                />
                <H3>Coming soon</H3>
                <Text>We are working on making this live.</Text>
            </div>
        </>
    )
}

function buildFrequencyDatum(
    datums: { daysUsed: number; frequency: number }[],
    min: number,
    max: number,
    isGradual = true
): FrequencyDatum[] {
    console.log('isGradual', isGradual)
    const result: FrequencyDatum[] = []
    for (let days = min; days <= max; ++days) {
        const index = datums.findIndex(datum => datum.daysUsed >= days)
        if (isGradual || days === max) {
            result.push({
                label: `${days} days`,
                value: index >= 0 ? datums.slice(index).reduce((sum, datum) => sum + datum.frequency, 0) : 0,
            })
        } else if (index >= 0 && datums[index].daysUsed === days) {
            result.push({
                label: `${days} days`,
                value: datums[index].frequency,
            })
        } else {
            result.push({
                label: `${days}+ days`,
                value: 0,
            })
        }
    }

    return result
}

function buildStandardDatum(datums: StandardDatum[], dateRange: AnalyticsDateRange): StandardDatum[] {
    // Generates 0 value series for dates that don't exist in the original data
    const [to, daysOffset] =
        dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
            ? [startOfWeek(new Date(), { weekStartsOn: 1 }), 7]
            : [startOfDay(new Date()), 1]
    const from =
        dateRange === AnalyticsDateRange.LAST_THREE_MONTHS
            ? sub(to, { months: 3 })
            : dateRange === AnalyticsDateRange.LAST_MONTH
            ? sub(to, { months: 1 })
            : sub(to, { weeks: 1 })
    const newDatums: StandardDatum[] = []
    let date = to
    while (date >= from) {
        const datum = datums?.find(datum => getDayOfYear(datum.date) === getDayOfYear(date))
        newDatums.push(datum ? { ...datum, date } : { date, value: 0 })
        date = addDays(date, -daysOffset)
    }

    return newDatums
}

export const AnalyticsSearchPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_WEEK)
    const { data, error, loading } = useQuery<SearchStatisticsResult, SearchStatisticsVariables>(SEARCH_STATISTICS, {
        variables: {
            dateRange,
        },
    })
    const [stats, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { searches, fileViews, fileOpens } = data.site.analytics.search
        const stats: Series<StandardDatum>[] = [
            {
                id: 'searches',
                name: eventAggregation === 'count' ? 'Searches' : 'Users searched',
                color: 'var(--cyan)',
                data: buildStandardDatum(
                    searches.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'fileViews',
                name: eventAggregation === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
                data: buildStandardDatum(
                    fileViews.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'fileOpens',
                name: eventAggregation === 'count' ? 'File opens' : 'Users opened files',
                color: 'var(--body-color)',
                data: buildStandardDatum(
                    fileOpens.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]

        const legends: ValueLegendListProps['items'] = [
            {
                value: searches.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Searches' : 'Users searched',
                color: 'var(--cyan)',
            },
            {
                value: fileViews.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'File views' : 'Users viewed files',
                color: 'var(--orange)',
            },
            {
                value: fileOpens.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'File opens' : 'Users opened files',
                color: 'var(--body-color)',
            },
        ]
        return [stats, legends]
    }, [data, eventAggregation, dateRange])

    const timeSavedStats = useMemo(() => {
        if (!data) {
            return []
        }
        const { searches, fileViews, fileOpens } = data.site.analytics.search

        return [
            {
                label: 'Searches,\nfile views\n& file opens',
                color: 'var(--purple)',
                minPerItem: 5,
                description:
                    'Each search or file view represents a developer solving a code use problem, getting information an active incident, or other use case. ',
                value: searches.summary.totalCount + fileViews.summary.totalCount + fileOpens.summary.totalCount,
            },
        ]
    }, [data])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Search</AnalyticsPageTitle>

            <Card className="p-3">
                <div className="d-flex justify-content-end align-items-stretch mb-2">
                    <HorizontalSelect<AnalyticsDateRange>
                        value={dateRange}
                        label="Date&nbsp;range"
                        onChange={setDateRange}
                        items={[
                            { value: AnalyticsDateRange.LAST_WEEK, label: 'Last week' },
                            { value: AnalyticsDateRange.LAST_MONTH, label: 'Last month' },
                            { value: AnalyticsDateRange.LAST_THREE_MONTHS, label: 'Last 3 months' },
                            { value: AnalyticsDateRange.CUSTOM, label: 'Custom (coming soon)', disabled: true },
                        ]}
                    />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {stats && (
                    <div>
                        <ChartContainer
                            title={eventAggregation === 'count' ? 'User activity by day' : 'Unique users by day'}
                            labelX="Time"
                            labelY={eventAggregation === 'count' ? 'User activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={stats} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-2">
                            <ToggleSelect<typeof eventAggregation>
                                selected={eventAggregation}
                                onChange={setEventAggregation}
                                items={[
                                    {
                                        tooltip: 'total # of actions triggered',
                                        label: 'Totals',
                                        value: 'count',
                                    },
                                    {
                                        tooltip: 'unique # of users triggered',
                                        label: 'Uniques',
                                        value: 'uniqueUsers',
                                    },
                                ]}
                            />
                        </div>
                    </div>
                )}
                <H3 className="my-3">Time saved</H3>
                {timeSavedStats?.map(timeSavedStatItem => (
                    <TimeSavedCalculator key={timeSavedStatItem.label} {...timeSavedStatItem} />
                ))}
            </Card>
        </>
    )
}

export const AnalyticsNotebooksPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('count')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_WEEK)
    const { data, error, loading } = useQuery<NotebooksStatisticsResult, NotebooksStatisticsVariables>(
        NOTEBOOKS_STATISTICS,
        {
            variables: {
                dateRange,
            },
        }
    )
    const [stats, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { creations, views, blockRuns } = data.site.analytics.notebooks
        const stats: Series<StandardDatum>[] = [
            {
                id: 'creations',
                name: eventAggregation === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
                data: buildStandardDatum(
                    creations.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
            {
                id: 'views',
                name: eventAggregation === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
                data: buildStandardDatum(
                    views.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]
        const legends: ValueLegendListProps['items'] = [
            {
                value: creations.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Notebooks created' : 'Users created notebooks',
                color: 'var(--cyan)',
            },
            {
                value: views.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Notebook views' : 'Users viewed notebooks',
                color: 'var(--orange)',
            },
            {
                value: blockRuns.summary[eventAggregation === 'count' ? 'totalCount' : 'totalUniqueUsers'],
                description: eventAggregation === 'count' ? 'Block runs' : 'Users ran blocks',
                color: 'var(--body-color)',
                position: 'right',
            },
        ]

        return [stats, legends]
    }, [data, dateRange, eventAggregation])

    const timeSavedStats = useMemo(() => {
        if (!data) {
            return []
        }
        const timeSavedStats = [
            {
                label: 'Views',
                color: 'var(--body-color)',
                minPerItem: 5,
                description:
                    'Notebooks reduce the time it takes to create living documentation and share it. Each notebook view accounts for time saved by both creators and consumers of notebooks.',
                value: data.site.analytics.notebooks.views.summary.totalCount,
            },
        ]
        return timeSavedStats
    }, [data])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Notebooks</AnalyticsPageTitle>

            <Card className="p-3 position-relative">
                <div className="d-flex justify-content-end align-items-stretch mb-2">
                    <HorizontalSelect<AnalyticsDateRange>
                        value={dateRange}
                        label="Date&nbsp;range"
                        onChange={setDateRange}
                        items={[
                            { value: AnalyticsDateRange.LAST_WEEK, label: 'Last week' },
                            { value: AnalyticsDateRange.LAST_MONTH, label: 'Last month' },
                            { value: AnalyticsDateRange.LAST_THREE_MONTHS, label: 'Last 3 months' },
                            { value: AnalyticsDateRange.CUSTOM, label: 'Custom (coming soon)', disabled: true },
                        ]}
                    />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {stats && (
                    <div>
                        <ChartContainer
                            title={eventAggregation === 'count' ? 'User activity by day' : 'Unique users by day'}
                            labelX="Time"
                            labelY={eventAggregation === 'count' ? 'User activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={stats} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-2">
                            <ToggleSelect<typeof eventAggregation>
                                selected={eventAggregation}
                                onChange={setEventAggregation}
                                items={[
                                    {
                                        tooltip: 'total # of actions triggered',
                                        label: 'Totals',
                                        value: 'count',
                                    },
                                    {
                                        tooltip: 'unique # of users triggered',
                                        label: 'Uniques',
                                        value: 'uniqueUsers',
                                    },
                                ]}
                            />
                        </div>
                    </div>
                )}
                <H3 className="my-3">Time saved</H3>
                {timeSavedStats?.map(timeSavedStatItem => (
                    <TimeSavedCalculator key={timeSavedStatItem.label} {...timeSavedStatItem} />
                ))}
            </Card>
        </>
    )
}

export const AnalyticsUsersPage: React.FunctionComponent<RouteComponentProps<{}>> = () => {
    const [eventAggregation, setEventAggregation] = useState<'count' | 'uniqueUsers'>('uniqueUsers')
    const [dateRange, setDateRange] = useState<AnalyticsDateRange>(AnalyticsDateRange.LAST_WEEK)
    const { data, error, loading } = useQuery<UsersStatisticsResult, UsersStatisticsVariables>(USERS_STATISTICS, {
        variables: {
            dateRange,
        },
    })
    const [frequencies, legends] = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const legends: ValueLegendListProps['items'] = [
            {
                value: users.activity.summary.totalUniqueUsers,
                description: 'Active users',
                color: 'var(--purple)',
            },
            {
                value: data.users.totalCount,
                description: 'Registered Users',
                color: 'var(--body-color)',
                position: 'right',
            },
            {
                value: data.site.productSubscription.license?.userCount ?? 0,
                description: 'Users licenses',
                color: 'var(--body-color)',
                position: 'right',
            },
        ]

        const frequencies: FrequencyDatum[] = buildFrequencyDatum(users.frequencies, 1, 30)

        return [frequencies, legends]
    }, [data])

    const activities = useMemo(() => {
        if (!data) {
            return []
        }
        const { users } = data.site.analytics
        const activities: Series<StandardDatum>[] = [
            {
                id: 'activity',
                name: eventAggregation === 'count' ? 'Activities' : 'Active users',
                color: eventAggregation === 'count' ? 'var(--cyan)' : 'var(--purple)',
                data: buildStandardDatum(
                    users.activity.nodes.map(node => ({
                        date: new Date(node.date),
                        value: node[eventAggregation],
                    })),
                    dateRange
                ),
                getXValue: ({ date }) => date,
                getYValue: ({ value }) => value,
            },
        ]

        return activities
    }, [data, eventAggregation, dateRange])

    if (error) {
        throw error
    }

    if (loading) {
        return <LoadingSpinner />
    }

    return (
        <>
            <AnalyticsPageTitle>Analytics / Users</AnalyticsPageTitle>
            <Card className="p-3">
                <div className="d-flex justify-content-end align-items-stretch mb-2">
                    <HorizontalSelect<AnalyticsDateRange>
                        label="Date&nbsp;range"
                        value={dateRange}
                        onChange={setDateRange}
                        items={[
                            { value: AnalyticsDateRange.LAST_WEEK, label: 'Last week' },
                            { value: AnalyticsDateRange.LAST_MONTH, label: 'Last month' },
                            { value: AnalyticsDateRange.LAST_THREE_MONTHS, label: 'Last 3 months' },
                            { value: AnalyticsDateRange.CUSTOM, label: 'Custom (coming soon)', disabled: true },
                        ]}
                    />
                </div>
                {legends && <ValueLegendList className="mb-3" items={legends} />}
                {activities && (
                    <div>
                        <ChartContainer
                            title={eventAggregation === 'count' ? 'User activity by day' : 'Unique users by day'}
                            labelX="Time"
                            labelY={eventAggregation === 'count' ? 'User activity' : 'Unique users'}
                        >
                            {width => <LineChart width={width} height={300} series={activities} />}
                        </ChartContainer>
                        <div className="d-flex justify-content-end align-items-stretch mb-2">
                            <ToggleSelect<typeof eventAggregation>
                                selected={eventAggregation}
                                onChange={setEventAggregation}
                                items={[
                                    {
                                        tooltip: 'total # of actions triggered',
                                        label: 'Totals',
                                        value: 'count',
                                    },
                                    {
                                        tooltip: 'unique # of users triggered',
                                        label: 'Uniques',
                                        value: 'uniqueUsers',
                                    },
                                ]}
                            />
                        </div>
                    </div>
                )}
                {frequencies && (
                    <ChartContainer title="Frequency of use" labelX="Days used" labelY="Users">
                        {width => (
                            <BarChart
                                width={width}
                                height={300}
                                data={frequencies}
                                getDatumName={datum => datum.label}
                                getDatumValue={datum => datum.value}
                                getDatumColor={() => 'var(--oc-blue-2)'}
                            />
                        )}
                    </ChartContainer>
                )}
            </Card>
        </>
    )
}
