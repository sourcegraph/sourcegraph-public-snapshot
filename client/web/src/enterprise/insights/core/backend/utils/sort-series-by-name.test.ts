import { sortSeriesByName, SortSeriesByNameParameter } from './sort-series-by-name'

const toName = ({ name }: SortSeriesByNameParameter) => name

describe('sortSeriesByName', () => {
    it('sorts alphabetically', () => {
        const testSeries: SortSeriesByNameParameter[] = [
            {
                name: 'A',
            },
            {
                name: 'C',
            },
            {
                name: 'B',
            },
        ]

        expect(testSeries.sort(sortSeriesByName).map(toName)).toStrictEqual(['A', 'B', 'C'])
    })

    it('sorts on version', () => {
        const testSeries: SortSeriesByNameParameter[] = [
            {
                name: 'v2.0.1',
            },
            {
                name: 'v1.0.0',
            },
            {
                name: '1.2.3',
            },
        ]

        expect(testSeries.sort(sortSeriesByName).map(toName)).toStrictEqual(['v1.0.0', '1.2.3', 'v2.0.1'])
    })

    it('sorts on version and alphabetically', () => {
        const testSeries: SortSeriesByNameParameter[] = [
            {
                name: 'B',
            },
            {
                name: 'v2.0.1',
            },
            {
                name: 'v1.0.0',
            },
            {
                name: 'C',
            },
            {
                name: '1.2.3',
            },
            {
                name: 'A',
            },
        ]

        expect(testSeries.sort(sortSeriesByName).map(toName)).toStrictEqual([
            'A',
            'B',
            'C',
            'v1.0.0',
            '1.2.3',
            'v2.0.1',
        ])
    })
})
