import { SeriesType, SeriesWithData } from './data-series-processing'
import { getMinMaxBoundaries } from './get-min-max-boundary'

interface Datum {
    x: Date
    a: number | null
    b: number | null
    c: number | null
}

const testSeriesWithData: SeriesWithData<Datum>[] = [
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
]

describe('getMinMaxBoundary', () => {
    it('should calculate a valid boundary box', () => {
        expect(getMinMaxBoundaries({ dataSeries: testSeriesWithData, zeroYAxisMin: false })).toStrictEqual({
            minX: +new Date(2022, 2, 2),
            minY: 1,
            maxX: +new Date(2022, 2, 6),
            maxY: 3,
        })
    })

    it('should calculate a valid boundary box with zeroYAxisMin setting', () => {
        expect(getMinMaxBoundaries({ dataSeries: testSeriesWithData, zeroYAxisMin: true })).toStrictEqual({
            minX: +new Date(2022, 2, 2),
            minY: 0,
            maxX: +new Date(2022, 2, 6),
            maxY: 3,
        })
    })
})
