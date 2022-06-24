/* eslint-disable react/forbid-dom-props */
import { useMemo, useState } from 'react'

import { sub } from 'date-fns'
import { RouteComponentProps } from 'react-router'

import { H2, Card, Tabs, TabList, Tab, TabPanels, TabPanel, Select, Input, H3, Text } from '@sourcegraph/wildcard'

import { LineChart, ParentSize, Series } from '../../charts'
import { PageTitle } from '../../components/PageTitle'

import styles from './index.module.scss'

interface CalculatorProps {
    color: string
    name: string
    count: number
    savedPerCount: number
    description?: string
}

const Calculator: React.FunctionComponent<CalculatorProps> = ({ color, count, description, savedPerCount, name }) => {
    const [minutesPerCount, setMinutesPerCount] = useState(savedPerCount)
    const hoursSaved = useMemo(() => (count * minutesPerCount) / 60, [count, minutesPerCount])
    return (
        <Card className="mb-3 p-2 d-flex justify-content-between flex-row" key={name}>
            <Text className="mr-3">
                <Text
                    style={{
                        color,
                    }}
                    className={styles.count}
                >
                    {count}
                </Text>
                {name}
            </Text>
            <Text className="mr-3">
                <Input
                    type="number"
                    value={minutesPerCount}
                    onChange={event => setMinutesPerCount(event.target.value)}
                />
                minutes saved
            </Text>
            <Text className="mr-3">
                <Text className={styles.count}>{hoursSaved.toFixed(1)}</Text>
                hours saved
            </Text>
            <Text className="flex-1">
                <Text className="font-weight-bold">About this statistics</Text>
                {description}
            </Text>
        </Card>
    )
}

interface StandardDatum {
    date: Date
    value: number
}
interface ChartDataItem {
    totalCount: number
    name: string
    color: string
    series: StandardDatum[]
    showDevTimeCalculator?: boolean
    description?: string
}
interface ChartProps {
    onDateRangeChange: (dateRange: DateRange) => void
    dateRange: DateRange
    data: ChartDataItem[]
}

const Chart: React.FunctionComponent<ChartProps> = ({ data, dateRange, onDateRangeChange }) => {
    const series: Series<StandardDatum>[] = useMemo(
        () =>
            data.map(item => ({
                id: item.name,
                name: item.name,
                data: item.series,
                color: item.color,
                getXValue: datum => datum.date,
                getYValue: datum => datum.value,
            })),
        [data]
    )
    return (
        <Card className="p-2">
            <div className="d-flex justify-content-end">
                <Select
                    id="date-range"
                    label="Date&nbsp;range"
                    isCustomStyle={true}
                    className="d-flex align-items-baseline"
                    selectClassName="ml-2"
                    onChange={value => onDateRangeChange(value.target.value as DateRange)}
                >
                    {Object.entries(DateRange).map(([key, value]) => (
                        <option key={key} value={value} selected={dateRange === value}>
                            {value}
                        </option>
                    ))}
                    <option value="custom" disabled={true}>
                        Custom (coming soon)
                    </option>
                </Select>
            </div>
            <div className="d-flex justify-content-left">
                {data.map(item => (
                    <div key={item.name} className="d-flex flex-column align-items-center mr-3">
                        <span style={{ color: item.color }} className={styles.count}>
                            {item.totalCount}
                        </span>
                        {item.name}
                    </div>
                ))}
            </div>
            <ParentSize>{({ width }) => <LineChart width={width} height={400} series={series} />}</ParentSize>
            <H3 className="m-3">Time saved</H3>
            {data.map(item => (
                <Calculator
                    key={item.name}
                    color={item.color}
                    name={item.name}
                    savedPerCount={5}
                    description={item.description}
                    count={item.totalCount}
                />
            ))}
        </Card>
    )
}

enum DateRange {
    LastThreeMonths = 'Last 3 months',
    LastMonth = 'Last month',
    LastWeek = 'Last week',
}

const StatisticItem: React.FunctionComponent<{
    title: string
    items: Omit<ChartDataItem, 'series' | 'totalCount'>[]
}> = ({ title, items }) => {
    const [dateRange, setDateRange] = useState<DateRange>(DateRange.LastWeek)
    const data = useMemo(() => {
        const now = new Date()
        const fromDate =
            dateRange === DateRange.LastThreeMonths
                ? sub(now, { months: 3 })
                : dateRange === DateRange.LastMonth
                ? sub(now, { months: 1 })
                : sub(now, { weeks: 1 })
        return getMockData(fromDate, now, items)
    }, [dateRange, items])
    return (
        <>
            <H2 className="mt-4"> {title}</H2>
            <Chart data={data} dateRange={dateRange} onDateRangeChange={setDateRange} />
        </>
    )
}

export const AdvancedStatisticsPage: React.FunctionComponent<RouteComponentProps<{}>> = () => (
    <>
        <PageTitle title="Usage statistics - Admin" />
        <Tabs lazy={true} behavior="memoize" size="large">
            <TabList>
                {/* <Tab>Overview</Tab> */}
                <Tab>Search</Tab>
                <Tab>Code intel</Tab>
                <Tab>Users</Tab>
                <Tab>Code insights</Tab>
                <Tab>Batch changes</Tab>
                <Tab>Notebooks</Tab>
                <Tab>Extensions</Tab>
            </TabList>
            <TabPanels>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Search"
                        items={[
                            {
                                name: 'Searches',
                                color: 'var(--cyan)',
                            },
                            {
                                name: 'File views',
                                color: 'var(--orange)',
                                showDevTimeCalculator: true,
                                description:
                                    'Notebooks save developers time by reducing the time required to find, read, and understand code. Enter the minutes saved per view to ballpark developer hours saved. ',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Code intel"
                        items={[
                            {
                                name: 'References',
                                color: 'var(--cyan)',
                            },
                            {
                                name: 'Definitions',
                                color: 'var(--orange)',
                                showDevTimeCalculator: true,
                                description:
                                    'Notebooks save developers time by reducing the time required to find, read, and understand code. Enter the minutes saved per view to ballpark developer hours saved. ',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Users"
                        items={[
                            {
                                name: 'Total users',
                                color: 'var(--purple)',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Code insights"
                        items={[
                            {
                                name: 'Insights created',
                                color: 'var(--cyan)',
                            },
                            {
                                name: 'Insights viewed',
                                color: 'var(--orange)',
                                showDevTimeCalculator: true,
                                description:
                                    'Notebooks save developers time by reducing the time required to find, read, and understand code. Enter the minutes saved per view to ballpark developer hours saved. ',
                            },
                            {
                                name: 'Datapoint clicked',
                                color: 'var(--purple)',
                                showDevTimeCalculator: true,
                                description:
                                    'Notebooks save developers time by reducing the time required to find, read, and understand code. Enter the minutes saved per view to ballpark developer hours saved. ',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Batch changes"
                        items={[
                            {
                                name: 'Changesets created',
                                color: 'var(--blue)',
                            },
                            {
                                name: 'Changesets merged',
                                color: 'var(--cyan)',
                                showDevTimeCalculator: true,
                                description:
                                    'Notebooks save developers time by reducing the time required to find, read, and understand code. Enter the minutes saved per view to ballpark developer hours saved. ',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Notebooks"
                        items={[
                            {
                                name: 'Notebooks created',
                                color: 'var(--cyan)',
                            },
                            {
                                name: 'Notebooks viewed',
                                color: 'var(--purple)',
                                showDevTimeCalculator: true,
                                description:
                                    'Notebooks save developers time by reducing the time required to find, read, and understand code. Enter the minutes saved per view to ballpark developer hours saved. ',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>
                    <StatisticItem
                        title="Statistics / Extensions"
                        items={[
                            {
                                name: 'Extension uses',
                                color: 'var(--orange)',
                            },
                        ]}
                    />
                </TabPanel>
                <TabPanel>Coming soon</TabPanel>
            </TabPanels>
        </Tabs>
    </>
)

const getMockData = (
    fromDate: Date,
    toDate: Date,
    items: Omit<ChartDataItem, 'series' | 'totalCount'>[]
): ChartDataItem[] =>
    items
        .map(item => ({ ...item, series: generateRandomDataSeries(fromDate, toDate) }))
        .map(item => ({
            ...item,
            totalCount: item.series.map(item => item.value).reduce((a, b) => a + b, 0),
        })) as ChartDataItem[]

function generateRandomDataSeries(fromDate: Date, toDate: Date): StandardDatum[] {
    const randomData: StandardDatum[] = []
    const days = Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24))
    for (let index = 0; index < days; index++) {
        randomData.push({
            date: new Date(fromDate.getTime() + index * 1000 * 60 * 60 * 24),
            value: Math.floor(Math.random() * 90) + 10,
        })
    }

    return randomData
}
