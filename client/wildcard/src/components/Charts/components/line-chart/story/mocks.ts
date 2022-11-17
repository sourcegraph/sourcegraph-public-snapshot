import { Series } from '../../../types'

export interface StandardDatum {
    value: number
    x: Date | number
    link?: string
}

const getXValue = (datum: StandardDatum): Date => new Date(datum.x)
const getYValue = (datum: StandardDatum): number => datum.value
const getLinkURL = (datum: StandardDatum): string | undefined => datum.link

export const STANDARD_SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: new Date(2020, 12, 25), value: 12, link: 'https://google.com/search' },
            { x: new Date(2021, 2, 25), value: 17, link: 'https://google.com/search' },
            { x: new Date(2021, 4, 25), value: 19, link: 'https://google.com/search' },
            { x: new Date(2021, 6, 25), value: 23, link: 'https://google.com/search' },
            { x: new Date(2021, 8, 25), value: 27 },
            { x: new Date(2021, 10, 25), value: 28 },
            { x: new Date(2021, 12, 25), value: 29 },
            { x: new Date(2022, 2, 25), value: 30 },
            { x: new Date(2022, 4, 25), value: 31 },
            { x: new Date(2022, 6, 25), value: 33 },
            { x: new Date(2022, 8, 25), value: 36 },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue,
        getYValue,
        getLinkURL,
    },
    {
        id: 'series_002',
        data: [
            { x: new Date(2020, 12, 25), value: 9, link: 'https://twitter.com/search' },
            { x: new Date(2021, 2, 25), value: 10, link: 'https://twitter.com/search' },
            { x: new Date(2021, 4, 25), value: 12, link: 'https://twitter.com/search' },
            { x: new Date(2021, 6, 25), value: 16, link: 'https://twitter.com/search' },
            { x: new Date(2021, 8, 25), value: 19 },
            { x: new Date(2021, 10, 25), value: 22 },
            { x: new Date(2021, 12, 25), value: 25 },
            { x: new Date(2022, 2, 25), value: 26 },
            { x: new Date(2022, 4, 25), value: 26 },
            { x: new Date(2022, 6, 25), value: 29 },
            { x: new Date(2022, 8, 25), value: 31 },
        ],
        name: 'C metric',
        color: 'var(--green)',
        getXValue,
        getYValue,
        getLinkURL,
    },
    {
        id: 'series_003',
        data: [
            { x: new Date(2020, 12, 25), value: 8, link: 'https://yandex.com/search' },
            { x: new Date(2021, 2, 25), value: 13, link: 'https://yandex.com/search' },
            { x: new Date(2021, 4, 25), value: 22, link: 'https://yandex.com/search' },
            { x: new Date(2021, 6, 25), value: 23, link: 'https://yandex.com/search' },
            { x: new Date(2021, 8, 25), value: 24 },
            { x: new Date(2021, 10, 25), value: 20 },
            { x: new Date(2021, 12, 25), value: 17 },
            { x: new Date(2022, 2, 25), value: 17 },
            { x: new Date(2022, 4, 25), value: 18 },
            { x: new Date(2022, 6, 25), value: 21 },
            { x: new Date(2022, 8, 25), value: 26 },
        ],
        name: 'B metric',
        color: 'var(--yellow)',
        getXValue,
        getYValue,
        getLinkURL,
    },
]

/**
 * Example dataset that has segment where all three lines overlap.
 */
export const FLAT_SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: new Date(2020, 12, 25), value: 12, link: 'https://google.com/search' },
            { x: new Date(2021, 2, 25), value: 17, link: 'https://google.com/search' },
            { x: new Date(2021, 4, 25), value: 15, link: 'https://google.com/search' },
            { x: new Date(2021, 6, 25), value: 15, link: 'https://google.com/search' },
            { x: new Date(2021, 8, 25), value: 15 },
            { x: new Date(2021, 10, 25), value: 15 },
            { x: new Date(2021, 12, 25), value: 29 },
            { x: new Date(2022, 2, 25), value: 30 },
            { x: new Date(2022, 4, 25), value: 31 },
            { x: new Date(2022, 6, 25), value: 33 },
            { x: new Date(2022, 8, 25), value: 36 },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue,
        getYValue,
        getLinkURL,
    },
    {
        id: 'series_002',
        data: [
            { x: new Date(2020, 12, 25), value: 9, link: 'https://twitter.com/search' },
            { x: new Date(2021, 2, 25), value: 10, link: 'https://twitter.com/search' },
            { x: new Date(2021, 4, 25), value: 15, link: 'https://twitter.com/search' },
            { x: new Date(2021, 6, 25), value: 15, link: 'https://twitter.com/search' },
            { x: new Date(2021, 8, 25), value: 15 },
            { x: new Date(2021, 10, 25), value: 15 },
            { x: new Date(2021, 12, 25), value: 25 },
            { x: new Date(2022, 2, 25), value: 26 },
            { x: new Date(2022, 4, 25), value: 26 },
            { x: new Date(2022, 6, 25), value: 29 },
            { x: new Date(2022, 8, 25), value: 31 },
        ],
        name: 'C metric',
        color: 'var(--green)',
        getXValue,
        getYValue,
        getLinkURL,
    },
    {
        id: 'series_003',
        data: [
            { x: new Date(2020, 12, 25), value: 8, link: 'https://yandex.com/search' },
            { x: new Date(2021, 2, 25), value: 13, link: 'https://yandex.com/search' },
            { x: new Date(2021, 4, 25), value: 15, link: 'https://yandex.com/search' },
            { x: new Date(2021, 6, 25), value: 15, link: 'https://yandex.com/search' },
            { x: new Date(2021, 8, 25), value: 15 },
            { x: new Date(2021, 10, 25), value: 15 },
            { x: new Date(2021, 12, 25), value: 17 },
            { x: new Date(2022, 2, 25), value: 17 },
            { x: new Date(2022, 4, 25), value: 18 },
            { x: new Date(2022, 6, 25), value: 21 },
            { x: new Date(2022, 8, 25), value: 26 },
        ],
        name: 'B metric',
        color: 'var(--yellow)',
        getXValue,
        getYValue,
        getLinkURL,
    },
]

/**
 * Datasets where series have big values (in order to test label overflow logic)
 */
export const SERIES_WITH_HUGE_DATA: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: new Date(2022, 1), value: 95_000 },
            { x: new Date(2022, 2), value: 125_000 },
            { x: new Date(2022, 3), value: 195_000 },
            { x: new Date(2022, 4), value: 235_000 },
            { x: new Date(2022, 5), value: 325_000 },
            { x: new Date(2022, 6), value: 400_000 },
            { x: new Date(2022, 7), value: 520_000 },
            { x: new Date(2022, 8), value: 720_000 },
            { x: new Date(2022, 9), value: 780_000 },
            { x: new Date(2022, 10), value: 800_000 },
            { x: new Date(2022, 11), value: 815_000 },
            { x: new Date(2022, 12), value: 840_000 },
        ],
        name: 'Fix',
        color: 'var(--oc-indigo-7)',
        getXValue,
        getYValue,
    },
    {
        id: 'series_002',
        data: [
            { x: new Date(2022, 1), value: 92_000 },
            { x: new Date(2022, 2), value: 100_000 },
            { x: new Date(2022, 3), value: 106_000 },
            { x: new Date(2022, 4), value: 120_000 },
            { x: new Date(2022, 5), value: 130_000 },
            { x: new Date(2022, 6), value: 136_000 },
        ],
        color: 'var(--oc-orange-7)',
        name: 'Revert',
        getXValue,
        getYValue,
    },
]

export const UNALIGNED_SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 94 },
            { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 134 },
            { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 134 },
            { x: 1588965700286, value: 123 },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue,
        getYValue,
    },
    {
        id: 'series_002',
        data: [
            { x: 1588965700286 - 1.4 * 24 * 60 * 60 * 1000, value: 150 },
            { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: 150 },
            { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 190 },
            { x: 1588965700286, value: 170 },
            { x: 1588965700286 + 24 * 60 * 60 * 1000, value: 200 },
            { x: 1588965700286 + 1.3 * 24 * 60 * 60 * 1000, value: 180 },
        ],
        name: 'C metric',
        color: 'var(--purple)',
        getXValue,
        getYValue,
    },
    {
        id: 'series_003',
        data: [
            { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, value: 200 },
            { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, value: 150 },
            { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 190 },
            { x: 1588965700286, value: 170 },
        ],
        name: 'B metric',
        color: 'var(--warning)',
        getXValue,
        getYValue,
    },
]
