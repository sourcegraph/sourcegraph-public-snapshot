import { describe, expect, test } from 'vitest'

import { decodeSearchInsightUrl, encodeSearchInsightUrl } from './search-insight-url-parsers'

describe('decodeSearchInsightUrl', () => {
    test('should return null of non of insight relevant fields are presented in the URL query params', () => {
        const queryString = encodeURIComponent('?hell=there')

        expect(decodeSearchInsightUrl(queryString)).toBe(null)
    })

    test('should return a valid search insight initial values object', () => {
        const queryString = encodeURIComponent(
            `?repositories=github.com/sourcegraph/sourcegraph,github.com/example/example&title=Insight title&series=${JSON.stringify(
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

        expect(decodeSearchInsightUrl(queryString)).toStrictEqual({
            repoMode: 'urls-list',
            repoQuery: { query: '' },
            repositories: ['github.com/sourcegraph/sourcegraph', 'github.com/example/example'],
            title: 'Insight title',
            series: [
                { id: 1, edit: false, valid: true, autofocus: false, name: 'series 1', query: 'test1', stroke: 'red' },
                { id: 2, edit: false, valid: true, autofocus: false, name: 'series 2', query: 'test2', stroke: 'blue' },
            ],
        })
    })
})

describe('encodeSearchInsightUrl', () => {
    test('should encode search insight values in a way that they could be decoded with decodeUrlSearchInsight', () => {
        const encodedSearchInsightParameters = encodeSearchInsightUrl({
            repositories: ['github.com/sourcegraph/sourcegraph', 'github.com/example/example'],
            title: 'Insight title',
            series: [
                { id: '1', name: 'series 1', query: 'test1', stroke: 'red' },
                { id: '2', name: 'series 2', query: 'test2', stroke: 'blue' },
            ],
        })

        expect(decodeSearchInsightUrl(encodedSearchInsightParameters)).toStrictEqual({
            repoMode: 'urls-list',
            repoQuery: { query: '' },
            repositories: ['github.com/sourcegraph/sourcegraph', 'github.com/example/example'],
            title: 'Insight title',
            series: [
                {
                    id: '1',
                    edit: false,
                    valid: true,
                    autofocus: false,
                    name: 'series 1',
                    query: 'test1',
                    stroke: 'red',
                },
                {
                    id: '2',
                    edit: false,
                    valid: true,
                    autofocus: false,
                    name: 'series 2',
                    query: 'test2',
                    stroke: 'blue',
                },
            ],
        })
    })
})
