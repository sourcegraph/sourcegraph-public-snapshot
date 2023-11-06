import { describe, expect, test } from '@jest/globals'

import { ExternalServiceKind } from '../../graphql-operations'

import { getLinksFromString } from './get-links'

const externalURL: { url: string; serviceKind: ExternalServiceKind | null } = {
    url: 'https://github.com/sourcegraph/sourcegraph',
    serviceKind: ExternalServiceKind.GITHUB,
}

describe('get-links', () => {
    test('parses urls and GitHub issues', () => {
        const example = 'This contains a url https://sourcegraph.com. This contains a GH issue #1234'
        const result = getLinksFromString({ input: example, externalURLs: [externalURL] })
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "end": 43,
                "href": "https://sourcegraph.com",
                "start": 20,
                "type": "url",
                "value": "https://sourcegraph.com",
              },
              Object {
                "end": 75,
                "href": "https://github.com/sourcegraph/sourcegraph/pull/1234",
                "start": 70,
                "type": "gh-issue",
                "value": "#1234",
              },
            ]
        `)
    })

    test('parses overlapping URLs and GitHub issues', () => {
        const example = 'This contains a URL that could be mistaken for a GH issue https://sourcegraph.com/(#1234)'
        const result = getLinksFromString({
            input: example,
            externalURLs: [externalURL],
        })
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "end": 89,
                "href": "https://sourcegraph.com/(#1234)",
                "start": 58,
                "type": "url",
                "value": "https://sourcegraph.com/(#1234)",
              },
            ]
        `)
    })

    test('does not parse GitHub issues if no external URLS', () => {
        const example = 'This contains a GH issue #1234'
        const result = getLinksFromString({
            input: example,
        })
        expect(result).toHaveLength(0)
    })

    test('does not parse file names', () => {
        const example = 'This contains a file name that could be mistaken for a URL: example/test/rust.rs'
        const result = getLinksFromString({
            input: example,
            externalURLs: [externalURL],
        })
        expect(result).toHaveLength(0)
    })
})
