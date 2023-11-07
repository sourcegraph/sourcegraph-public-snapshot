import { describe, expect, test } from '@jest/globals'

import { encodeCaptureInsightURL, decodeCaptureInsightURL } from './capture-insight-url-parsers'

describe('decodeCaptureInsightURL', () => {
    test('should return null of non of insight relevant fields are presented in the URL query params', () => {
        const queryString = encodeURIComponent('?hell=there')

        expect(decodeCaptureInsightURL(queryString)).toBe(null)
    })
})

describe('encodeSearchInsightUrl', () => {
    test('should encode search insight values in a way that they could be decoded with decodeUrlSearchInsight', () => {
        const encodedSearchInsightParameters = encodeCaptureInsightURL({
            repositories: ['github.com/sourcegraph/sourcegraph', 'github.com/example/example'],
            title: 'Insight title',
            groupSearchQuery: 'file:go\\.mod$ go\\s*(\\d\\.\\d+) patterntype:regexp',
        })

        expect(decodeCaptureInsightURL(encodedSearchInsightParameters)).toStrictEqual({
            repositories: ['github.com/sourcegraph/sourcegraph', 'github.com/example/example'],
            repoMode: 'urls-list',
            repoQuery: { query: '' },
            title: 'Insight title',
            groupSearchQuery: 'file:go\\.mod$ go\\s*(\\d\\.\\d+) patterntype:regexp',
        })
    })
})
