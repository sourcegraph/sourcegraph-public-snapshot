import { describe, expect, it } from '@jest/globals'

import type { Series } from '../../../../types'

import { getSeriesData } from './get-series-data'
import { SeriesType } from './types'

interface Datum {
    x: Date
    value: number
}

const getXValue = (datum: Datum): Date => datum.x
const getYValue = (datum: Datum): number => datum.value

const testSeries: Series<Datum>[] = [
    {
        id: 'series_001',
        data: [
            { x: new Date(2022, 2, 4), value: 1 },
            { x: new Date(2022, 2, 6), value: 2 },
        ],
        name: 'Series a',
        getXValue,
        getYValue,
    },
    {
        id: 'series_002',
        data: [
            { x: new Date(2022, 2, 3), value: 2 },
            { x: new Date(2022, 2, 4), value: 2 },
            { x: new Date(2022, 2, 6), value: 2 },
        ],
        name: 'Series b',
        getXValue,
        getYValue,
    },
    {
        id: 'series_003',
        data: [
            { x: new Date(2022, 2, 3), value: 3 },
            { x: new Date(2022, 2, 4), value: 3 },
            { x: new Date(2022, 2, 5), value: 3 },
        ],
        name: 'Series c',
        getXValue,
        getYValue,
    },
]

describe('getSeriesData', () => {
    it('should generate series with standard (non-stacked) datum list for each series', () => {
        expect(
            getSeriesData({
                series: testSeries,
                stacked: false,
            })
        ).toStrictEqual([
            {
                type: SeriesType.Independent,
                id: 'series_001',
                name: 'Series a',
                data: [
                    {
                        id: '0.series_001',
                        y: 1,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), value: 1 },
                    },
                    {
                        id: '1.series_001',
                        y: 2,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), value: 2 },
                    },
                ],
                getXValue,
                getYValue,
            },
            {
                type: SeriesType.Independent,
                id: 'series_002',
                name: 'Series b',
                data: [
                    {
                        id: '0.series_002',
                        y: 2,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), value: 2 },
                    },
                    {
                        id: '1.series_002',
                        y: 2,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), value: 2 },
                    },
                    {
                        id: '2.series_002',
                        y: 2,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), value: 2 },
                    },
                ],
                getXValue,
                getYValue,
            },
            {
                type: SeriesType.Independent,
                id: 'series_003',
                name: 'Series c',
                data: [
                    {
                        id: '0.series_003',
                        y: 3,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), value: 3 },
                    },
                    {
                        id: '1.series_003',
                        y: 3,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), value: 3 },
                    },
                    {
                        id: '2.series_003',
                        y: 3,
                        x: new Date(2022, 2, 5),
                        datum: { x: new Date(2022, 2, 5), value: 3 },
                    },
                ],
                getXValue,
                getYValue,
            },
        ])
    })

    it('should generate series with stacked datum list for each stacked series', () => {
        expect(
            getSeriesData({
                series: testSeries,
                stacked: true,
            })
        ).toStrictEqual([
            {
                type: SeriesType.Stacked,
                id: 'series_001',
                name: 'Series a',
                getXValue,
                getYValue,
                data: [
                    {
                        id: '0.series_001',
                        y0: 0,
                        y1: 1,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), value: 1 },
                    },
                    {
                        id: '1.series_001',
                        y0: 0,
                        y1: 2,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), value: 2 },
                    },
                ],
            },
            {
                type: SeriesType.Stacked,
                id: 'series_002',
                name: 'Series b',
                getXValue,
                getYValue,
                data: [
                    {
                        id: '0.series_002',
                        y0: 0,
                        y1: 2,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), value: 2 },
                    },
                    {
                        id: '1.series_002',
                        y0: 1,
                        y1: 3,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), value: 2 },
                    },
                    {
                        id: '2.series_002',
                        y0: 2,
                        y1: 4,
                        x: new Date(2022, 2, 6),
                        datum: { x: new Date(2022, 2, 6), value: 2 },
                    },
                ],
            },
            {
                type: SeriesType.Stacked,
                id: 'series_003',
                name: 'Series c',
                getXValue,
                getYValue,
                data: [
                    {
                        id: '0.series_003',
                        y0: 2,
                        y1: 5,
                        x: new Date(2022, 2, 3),
                        datum: { x: new Date(2022, 2, 3), value: 3 },
                    },
                    {
                        id: '1.series_003',
                        y0: 3,
                        y1: 6,
                        x: new Date(2022, 2, 4),
                        datum: { x: new Date(2022, 2, 4), value: 3 },
                    },
                    {
                        id: '2.series_003',
                        y0: 3.5,
                        y1: 6.5,
                        x: new Date(2022, 2, 5),
                        datum: { x: new Date(2022, 2, 5), value: 3 },
                    },
                ],
            },
        ])
    })
})
