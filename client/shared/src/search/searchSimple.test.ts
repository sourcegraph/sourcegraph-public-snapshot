import { describe, expect, test } from 'vitest'

import { hacksGobQueriesToRegex } from './searchSimple'

describe('hacksGobQueriesToRegex', () => {
    test('gob corner cases', () => {
        expect(hacksGobQueriesToRegex('')).toEqual('')

        // TODO should this match nothing? Right the filter is a noop
        expect(hacksGobQueriesToRegex('f:')).toEqual('f:')

        // TODO use a real parser
        // expect(hacksGobQueriesToRegex('f:"foo "')).toEqual('f:"foo "')

        // Match every repo and match every file act differently in our UX
        expect(hacksGobQueriesToRegex('r:*')).toEqual('r:')
        expect(hacksGobQueriesToRegex('repo:*')).toEqual('repo:')
        expect(hacksGobQueriesToRegex('f:*')).toEqual('f:.*')
        expect(hacksGobQueriesToRegex('file:*')).toEqual('file:.*')
    })

    test('converts repo filter to regex', () => {
        expect(hacksGobQueriesToRegex('repo:sourcegraph')).toEqual('repo:^sourcegraph$')
        expect(hacksGobQueriesToRegex('repo:github.com/*')).toEqual('repo:^github\\.com/')
        expect(hacksGobQueriesToRegex('repo:*/sourcegraph')).toEqual('repo:/sourcegraph$')
    })

    test('converts file filter to regex', () => {
        expect(hacksGobQueriesToRegex('file:README.md')).toEqual('file:^README\\.md$')
        expect(hacksGobQueriesToRegex('file:client/README.md')).toEqual('file:^client/README\\.md$')
        expect(hacksGobQueriesToRegex('file:*/Dockerfile')).toEqual('file:/Dockerfile$')
        expect(hacksGobQueriesToRegex('file:*.go')).toEqual('file:\\.go$')
        expect(hacksGobQueriesToRegex('file:src/*')).toEqual('file:^src/')
    })

    test('queries to regex', () => {
        expect(hacksGobQueriesToRegex('context:global repo:*/sourcegraph zoekt f:*.md')).toEqual(
            'context:global repo:/sourcegraph$ zoekt f:\\.md$'
        )
    })
})
