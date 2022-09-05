import { semanticSort } from './semantic-sort'

describe('sortSeriesByName', () => {
    it('sorts alphabetically', () => {
        const testSeries = ['A', 'C', 'B']

        expect(testSeries.sort(semanticSort)).toStrictEqual(['A', 'B', 'C'])
    })

    it('sorts on version', () => {
        const testSeries = ['v2.0.1', 'v1.0.0', '1.2.3']

        expect(testSeries.sort(semanticSort)).toStrictEqual(['v1.0.0', '1.2.3', 'v2.0.1'])
    })

    it('sorts on version and alphabetically', () => {
        const testSeries = ['B', 'v2.0.1', 'v1.0.0', 'C', '1.2.3', 'A']

        expect(testSeries.sort(semanticSort)).toStrictEqual(['A', 'B', 'C', 'v1.0.0', '1.2.3', 'v2.0.1'])
    })
})
