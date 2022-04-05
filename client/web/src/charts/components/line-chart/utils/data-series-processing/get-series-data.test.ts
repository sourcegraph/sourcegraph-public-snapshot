import { Series } from '../../../../types'

import { getSeriesData } from './get-series-data'
import { SeriesType } from './types'

interface Datum {
    x: Date
    a: number | null
    b: number | null
    c: number | null
}

const testDataList: Datum[] = [
    { x: new Date(2022, 2, 2), a: null, b: null, c: null },
    { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
    { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
    { x: new Date(2022, 2, 5), a: null, b: null, c: 3 },
    { x: new Date(2022, 2, 6), a: 2, b: 2, c: null },
]

const testSeries: Series<Datum>[] = [
    { dataKey: 'a', name: 'Series a' },
    { dataKey: 'b', name: 'Series b' },
    { dataKey: 'c', name: 'Series c' },
]

describe('getSeriesData', () => {
    it('should generate series with standard (non-stacked) datum list for each series', () => {
        expect(
            getSeriesData({
                data: testDataList,
                series: testSeries,
                stacked: false,
                getXValue: datum => datum.x,
            })
        ).toStrictEqual([
            {
                type: SeriesType.Independent,
                dataKey: 'a',
                name: 'Series a',
                data: [
                    {
                        y: null,
                        x: new Date(2022, 2, 2),
                        datum: { x: new Date(2022, 2, 2), a: null, b: null, c: null },
                    },
                    {
                        y: null,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
                    },
                    {
                        y: 1,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
                    },
                    {
                        y: 2,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), a: 2, b: 2, c: null },
                    },
                ],
            },
            {
                type: SeriesType.Independent,
                dataKey: 'b',
                name: 'Series b',
                data: [
                    {
                        y: null,
                        x: new Date(2022, 2, 2),
                        datum: { x: new Date(2022, 2, 2), a: null, b: null, c: null },
                    },
                    {
                        y: 2,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
                    },
                    {
                        y: 2,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
                    },
                    {
                        y: 2,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), a: 2, b: 2, c: null },
                    },
                ],
            },
            {
                type: SeriesType.Independent,
                dataKey: 'c',
                name: 'Series c',
                data: [
                    {
                        y: null,
                        x: new Date(2022, 2, 2),
                        datum: { x: new Date(2022, 2, 2), a: null, b: null, c: null },
                    },
                    {
                        y: 3,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
                    },
                    {
                        y: 3,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
                    },
                    {
                        y: 3,
                        x: new Date(2022, 2, 5),
                        datum: { x: new Date(2022, 2, 5), a: null, b: null, c: 3 },
                    },
                ],
            },
        ])
    })

    it('should generate series with stacked datum list for each stacked series', () => {
        expect(
            getSeriesData({
                data: testDataList,
                series: testSeries,
                stacked: true,
                getXValue: datum => datum.x,
            })
        ).toStrictEqual([
            {
                type: SeriesType.Stacked,
                dataKey: 'a',
                name: 'Series a',
                data: [
                    {
                        y0: null,
                        y1: null,
                        x: new Date(2022, 2, 2),
                        datum: { x: new Date(2022, 2, 2), a: null, b: null, c: null },
                    },
                    {
                        y0: null,
                        y1: null,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
                    },
                    {
                        y0: 0,
                        y1: 1,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
                    },
                    {
                        y0: 0,
                        y1: 2,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), a: 2, b: 2, c: null },
                    },
                ],
            },
            {
                type: SeriesType.Stacked,
                dataKey: 'b',
                name: 'Series b',
                data: [
                    {
                        y0: null,
                        y1: null,
                        x: new Date(2022, 2, 2),
                        datum: { x: new Date(2022, 2, 2), a: null, b: null, c: null },
                    },
                    {
                        y0: 0,
                        y1: 2,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
                    },
                    {
                        y0: 1,
                        y1: 3,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
                    },
                    {
                        y0: 2,
                        y1: 4,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), a: 2, b: 2, c: null },
                    },
                ],
            },
            {
                type: SeriesType.Stacked,
                dataKey: 'c',
                name: 'Series c',
                data: [
                    {
                        y0: null,
                        y1: null,
                        x: new Date(2022, 2, 2),
                        datum: { x: new Date(2022, 2, 2), a: null, b: null, c: null },
                    },
                    {
                        y0: 2,
                        y1: 5,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), a: null, b: 2, c: 3 },
                    },
                    {
                        y0: 3,
                        y1: 6,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), a: 1, b: 2, c: 3 },
                    },
                    {
                        y0: 3.5,
                        y1: 6.5,
                        x: new Date(2022, 2, 5),
                        datum: { x: new Date(2022, 2, 5), a: null, b: null, c: 3 },
                    },
                ],
            },
        ])
    })
})
