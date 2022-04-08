import { getFilteredSeriesData } from './helpers'
import { StandardSeriesDatum } from './types'

interface Datum {
    series1: number | null
    series2: number | null
}

describe('getFilteredSeriesData', () => {
    it("shouldn't modify data list without null gaps", () => {
        const testDataList: StandardSeriesDatum<Datum>[] = [
            { datum: { series1: 1, series2: 2 }, y: 1, x: new Date(2022, 2, 2) },
            { datum: { series1: 2, series2: 3 }, y: 2, x: new Date(2022, 2, 3) },
            { datum: { series1: 3, series2: 4 }, y: 3, x: new Date(2022, 2, 4) },
            { datum: { series1: 4, series2: 5 }, y: 4, x: new Date(2022, 2, 5) },
        ]

        expect(getFilteredSeriesData(testDataList)).toStrictEqual(testDataList)
    })

    it('should preserve null gaps at the beginning and remove all other null gaps', () => {
        const testDataList: StandardSeriesDatum<Datum>[] = [
            { datum: { series1: null, series2: null }, y: null, x: new Date(2022, 2, 2) },
            { datum: { series1: null, series2: null }, y: null, x: new Date(2022, 2, 3) },
            { datum: { series1: 3, series2: 4 }, y: 3, x: new Date(2022, 2, 4) },
            { datum: { series1: null, series2: null }, y: null, x: new Date(2022, 2, 5) },
            { datum: { series1: 4, series2: 5 }, y: 4, x: new Date(2022, 2, 6) },
        ]

        expect(getFilteredSeriesData(testDataList)).toStrictEqual([
            { datum: { series1: null, series2: null }, y: null, x: new Date(2022, 2, 2) },
            { datum: { series1: null, series2: null }, y: null, x: new Date(2022, 2, 3) },
            { datum: { series1: 3, series2: 4 }, y: 3, x: new Date(2022, 2, 4) },
            { datum: { series1: 4, series2: 5 }, y: 4, x: new Date(2022, 2, 6) },
        ])
    })
})
