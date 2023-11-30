import { describe, expect, it } from 'vitest'

import { SeriesType, type SeriesWithData } from './data-series-processing'
import { getMinMaxBoundaries } from './get-min-max-boundary'

interface Datum {
    x: Date
    value: number
}

const getXValue = (datum: Datum): Date => datum.x
const getYValue = (datum: Datum): number => datum.value

const testSeriesWithData: SeriesWithData<Datum>[] = [
    {
        type: SeriesType.Independent,
        id: 'a',
        name: 'Series a',
        getXValue,
        getYValue,
        data: [
            {
                id: 'a.1',
                y: 1,
                x: new Date(2022, 2, 4),
                datum: { x: new Date(2022, 2, 4), value: 1 },
            },
            {
                id: 'a.2',
                y: 2,
                x: new Date(2022, 2, 6),
                datum: { x: new Date(2022, 2, 6), value: 2 },
            },
        ],
    },
    {
        type: SeriesType.Independent,
        id: 'c',
        name: 'Series c',
        getXValue,
        getYValue,
        data: [
            {
                id: 'c.1',
                y: 3,
                x: new Date(2022, 2, 3),
                datum: { x: new Date(2022, 2, 3), value: 3 },
            },
            {
                id: 'c.2',
                y: 3,
                x: new Date(2022, 2, 4),
                datum: { x: new Date(2022, 2, 4), value: 3 },
            },
            {
                id: 'c.3',
                y: 3,
                x: new Date(2022, 2, 5),
                datum: { x: new Date(2022, 2, 5), value: 3 },
            },
        ],
    },
]

describe('getMinMaxBoundary', () => {
    it('should calculate a valid boundary box', () => {
        expect(getMinMaxBoundaries({ dataSeries: testSeriesWithData, zeroYAxisMin: false })).toStrictEqual({
            minX: +new Date(2022, 2, 3),
            minY: 1,
            maxX: +new Date(2022, 2, 6),
            maxY: 3,
        })
    })

    it('should calculate a valid boundary box with zeroYAxisMin setting', () => {
        expect(getMinMaxBoundaries({ dataSeries: testSeriesWithData, zeroYAxisMin: true })).toStrictEqual({
            minX: +new Date(2022, 2, 3),
            minY: 0,
            maxX: +new Date(2022, 2, 6),
            maxY: 3,
        })
    })
})
