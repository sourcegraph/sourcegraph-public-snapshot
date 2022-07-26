import { FC } from 'react'

import { ParentSize } from '@visx/responsive'

import { H4, Text } from '@sourcegraph/wildcard'

import { LineChart, Series } from '../../../charts'

import styles from './SearchInsightPanel.module.scss'

interface SearchInsightPanelProps {}

export const SearchInsightPanel: FC<SearchInsightPanelProps> = props => <HistoricalTrendInsight />

const HistoricalTrendInsight: FC<{}> = props => (
    <section>
        <H4>Historical trend</H4>
        <Text>...$ over last 30 days</Text>

        <ParentSize className={styles.chart}>
            {parent => (
                <>
                    <LineChart
                        width={parent.width}
                        height={parent.height}
                        series={MOCK_DATA}
                        className={styles.chartWithMock}
                    />

                    <Text className={styles.banner}>
                        Narrow your search down using the repo filter to see the historical trend.
                    </Text>
                </>
            )}
        </ParentSize>
    </section>
)

interface MockDatum {
    value: number | null
    x: Date
}

const MOCK_DATA: Series<MockDatum>[] = [
    {
        id: 'series_001',
        data: [
            {
                x: new Date(2020, 5, 5),
                value: 4000,
            },
            {
                x: new Date(2020, 5, 6),
                value: 4500,
            },
            {
                x: new Date(2020, 5, 7),
                value: 5500,
            },
            {
                x: new Date(2020, 5, 8),
                value: 6800,
            },
            {
                x: new Date(2020, 5, 9),
                value: 9000,
            },
            {
                x: new Date(2020, 5, 10),
                value: 10000,
            },
            {
                x: new Date(2020, 5, 11),
                value: 11000,
            },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue: datum => datum.x,
        getYValue: datum => datum.value,
    },
]
