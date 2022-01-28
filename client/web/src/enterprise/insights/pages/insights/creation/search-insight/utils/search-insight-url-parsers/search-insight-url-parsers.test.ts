import { decodeUrlSearchInsight, encodeUrlSearchInsight } from './search-insight-url-parsers'

describe('decodeUrlSearchInsight', () => {
    test('should return null of non of insight relevant fields are presented in the URL query params', () => {
        const queryString = encodeURIComponent('?hell=there')

        expect(decodeUrlSearchInsight(queryString)).toBe(null)
    })

    test('should return a valid search insight initial values object', () => {
        const queryString = encodeURIComponent(
            `?repositories=github.com/sourcegraph/sourcegraph, github.com/example/example&title=Insight title&allRepos=true&series=${JSON.stringify(
                [
                    {
                        id: 1,
                        name: 'series 1',
                        query: 'test1',
                        stroke: 'red',
                    },
                    { id: 2, name: 'series 2', query: 'test2', stroke: 'blue' },
                ]
            )}`.trim()
        )

        expect(decodeUrlSearchInsight(queryString)).toStrictEqual({
            repositories: 'github.com/sourcegraph/sourcegraph, github.com/example/example',
            title: 'Insight title',
            allRepos: true,
            series: [
                { id: 1, edit: true, valid: false, name: 'series 1', query: 'test1', stroke: 'red' },
                { id: 2, edit: true, valid: false, name: 'series 2', query: 'test2', stroke: 'blue' },
            ],
            step: 'days',
            stepValue: '8',
            visibility: '',
        })
    })
})

describe('encodeUrlSearchInsight', () => {
    test('should encode search insight values in a way that they could be decoded with decodeUrlSearchInsight', () => {
        const encodedSearchInsightParameters = encodeUrlSearchInsight({
            repositories: 'github.com/sourcegraph/sourcegraph, github.com/example/example',
            title: 'Insight title',
            allRepos: true,
            series: [
                { id: '1', name: 'series 1', query: 'test1', stroke: 'red' },
                { id: '2', name: 'series 2', query: 'test2', stroke: 'blue' },
            ],
        })

        expect(decodeUrlSearchInsight(encodedSearchInsightParameters)).toStrictEqual({
            repositories: 'github.com/sourcegraph/sourcegraph, github.com/example/example',
            title: 'Insight title',
            allRepos: true,
            series: [
                { id: '1', edit: true, valid: false, name: 'series 1', query: 'test1', stroke: 'red' },
                { id: '2', edit: true, valid: false, name: 'series 2', query: 'test2', stroke: 'blue' },
            ],
            step: 'days',
            stepValue: '8',
            visibility: '',
        })
    })
})
