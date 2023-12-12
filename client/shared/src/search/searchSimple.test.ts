import { describe, expect, test } from 'vitest'

import { hacksGobQueriesToRegex } from './searchSimple'

describe('hacksGobQueriesToRegex', () => {
    test('gob corner cases', () => {
        expect(hacksGobQueriesToRegex('')).toEqual('')

        // TODO should this match nothing? Right now the filter is a noop
        expect(hacksGobQueriesToRegex('f:')).toEqual('f:')

        // Quotes on regex filters just allow us to put spaces in
        expect(hacksGobQueriesToRegex('f:" "')).toEqual('f:"^ $"')
        expect(hacksGobQueriesToRegex('f:"* "')).toEqual('f:" $"')
        expect(hacksGobQueriesToRegex('f:" *"')).toEqual('f:"^ "')
        expect(hacksGobQueriesToRegex('f:"* *"')).toEqual('f:" "')
        expect(hacksGobQueriesToRegex('f:"foo bar"')).toEqual('f:"^foo bar$"')

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

    test('escaping', () => {
        expect(hacksGobQueriesToRegex('repo:\\*')).toEqual('repo:^\\*$')
        expect(hacksGobQueriesToRegex('repo:*\\*')).toEqual('repo:\\*$')
        expect(hacksGobQueriesToRegex('repo:\\**')).toEqual('repo:^\\*')
        expect(hacksGobQueriesToRegex('repo:*\\**')).toEqual('repo:\\*')

        expect(hacksGobQueriesToRegex('repo:foo\\*bar')).toEqual('repo:^foo\\*bar$')
        expect(hacksGobQueriesToRegex('repo:foo\\*')).toEqual('repo:^foo\\*$')
        expect(hacksGobQueriesToRegex('repo:\\*bar')).toEqual('repo:^\\*bar$')

        // Correctly escape regex char. We also don't double escape
        expect(hacksGobQueriesToRegex('repo:[')).toEqual('repo:^\\[$')
        expect(hacksGobQueriesToRegex('repo:\\[')).toEqual('repo:^\\[$')

        // TODO handle quoted strings
        // We escape the string twice, once for javascript interpreting this
        // string and once for the quotes inside of the string (which the
        // query layer parses)
        //expect(hacksGobQueriesToRegex('repo:"foo\\\\*bar"')).toEqual('repo:"^foo\\\\*bar$"')
        //expect(hacksGobQueriesToRegex('repo:"foo\\\\*"')).toEqual('repo:"^foo\\\\*$"')
        //expect(hacksGobQueriesToRegex('repo:"\\\\*bar"')).toEqual('repo:"^\\\\*bar$"')
    })

    test('queries to regex', () => {
        expect(hacksGobQueriesToRegex('context:global repo:*/sourcegraph zoekt f:*.md')).toEqual(
            'context:global repo:/sourcegraph$ zoekt f:\\.md$'
        )
    })
})
