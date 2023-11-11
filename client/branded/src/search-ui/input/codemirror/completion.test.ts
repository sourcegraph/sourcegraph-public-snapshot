import type { Completion } from '@codemirror/autocomplete'
import { describe, expect, test } from 'vitest'

import { SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { POPULAR_LANGUAGES } from '@sourcegraph/shared/src/search/query/languageFilter'
import { type ScanResult, scanSearchQuery, type ScanSuccess } from '@sourcegraph/shared/src/search/query/scanner'
import type { Token } from '@sourcegraph/shared/src/search/query/token'
import type { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { createDefaultSuggestionSources, FILTER_SHORTHAND_SUGGESTIONS, suggestionTypeFromTokens } from './completion'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

/**
 * Helper function for invoking all sources.
 */
async function getCompletionItems(
    token: Token,
    position: number,
    fetchSuggestions: () => Promise<SearchMatch[]>,
    options?: { isSourcegraphDotCom?: boolean; tokens?: Token[] }
) {
    const sources = createDefaultSuggestionSources({
        isSourcegraphDotCom: false,
        fetchSuggestions,
        ...options,
    })

    const results = await Promise.all(
        sources.map(source => source({ position, onAbort: () => {} }, options?.tokens || [token], token))
    )
    const allOptions: Completion[] = []
    for (const result of results) {
        if (result) {
            allOptions.push(...result.options)
        }
    }
    return allOptions
}

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

const getTokens = (query: string): Token[] => toSuccess(scanSearchQuery(query))
const getToken = (query: string, tokenIndex: number): Token => getTokens(query)[tokenIndex]
const shorthandSuggestions = FILTER_SHORTHAND_SUGGESTIONS.map(({ label }) => label)

// Using async as a short way to create functions that return promises
/* eslint-disable @typescript-eslint/require-await */
describe('codmirror completions', () => {
    test('returns static and dynamic filter type completions when the token matches a known filter', async () => {
        expect(
            (
                await getCompletionItems(getToken('re', 0), 2, async () => [
                    {
                        type: 'symbol',
                        repository: 'github.com/sourcegraph/jsonrpc2',
                        path: 'src/RepoRoutes.js',
                        symbols: [
                            {
                                kind: SymbolKind.VARIABLE,
                                name: 'RepoRoutes',
                                url: '',
                                containerName: '',
                                line: 1,
                            },
                        ],
                    },
                ])
            )?.map(({ label }) => label)
        ).toStrictEqual(
            [
                'after',
                'archived',
                'author',
                '-author',
                'before',
                'case',
                'committer',
                '-committer',
                'content',
                '-content',
                'context',
                'count',
                'file',
                '-file',
                'fork',
                'lang',
                '-lang',
                'message',
                '-message',
                'patterntype',
                'repo',
                '-repo',
                'repohascommitafter',
                'repohasfile',
                '-repohasfile',
                'rev',
                'select',
                'timeout',
                'type',
                'visibility',
                ...shorthandSuggestions,
                'variable RepoRoutes',
            ].filter(label => label.toLowerCase().includes('re'))
        )
    })

    test('returns suggestions for an empty query', async () => {
        expect((await getCompletionItems(getToken('', 0), 0, async () => []))?.map(({ label }) => label)).toStrictEqual(
            [
                'after',
                'archived',
                'author',
                '-author',
                'before',
                'case',
                'committer',
                '-committer',
                'content',
                '-content',
                'context',
                'count',
                'file',
                '-file',
                'fork',
                'lang',
                '-lang',
                'message',
                '-message',
                'patterntype',
                'repo',
                '-repo',
                'repohascommitafter',
                'repohasfile',
                '-repohasfile',
                'rev',
                'select',
                'timeout',
                'type',
                'visibility',
                ...shorthandSuggestions,
            ]
        )
    })

    test('returns suggestions on whitespace', async () => {
        expect(
            (await getCompletionItems(getToken('a ', 1), 2, async () => []))?.map(({ label }) => label)
        ).toStrictEqual([
            'after',
            'archived',
            'author',
            '-author',
            'before',
            'case',
            'committer',
            '-committer',
            'content',
            '-content',
            'context',
            'count',
            'file',
            '-file',
            'fork',
            'lang',
            '-lang',
            'message',
            '-message',
            'patterntype',
            'repo',
            '-repo',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
            ...shorthandSuggestions,
        ])
    })

    test('returns static filter type completions for case-insensitive query', async () => {
        expect(
            (await getCompletionItems(getToken('rE', 0), 2, async () => []))?.map(({ label }) => label)
        ).toStrictEqual(
            [
                'after',
                'archived',
                'author',
                '-author',
                'before',
                'case',
                'committer',
                '-committer',
                'content',
                '-content',
                'context',
                'count',
                'file',
                '-file',
                'fork',
                'lang',
                '-lang',
                'message',
                '-message',
                'patterntype',
                'repo',
                '-repo',
                'repohascommitafter',
                'repohasfile',
                '-repohasfile',
                'rev',
                'select',
                'timeout',
                'type',
                'visibility',
                ...shorthandSuggestions,
            ].filter(label => label.toLowerCase().includes('re'))
        )
    })

    test('returns completions for filters with discrete values', async () => {
        expect(
            (await getCompletionItems(getToken('case:y', 0), 6, async () => []))?.map(({ label }) => label)
        ).toStrictEqual(['yes', 'no'])
    })

    test('returns completions for filters with static suggestions', async () => {
        expect(
            (await getCompletionItems(getToken('lang:', 0), 5, async () => []))?.map(({ label }) => label)
        ).toStrictEqual(POPULAR_LANGUAGES)
    })

    test('returns completions in order of discrete value definition, not alphabetically', async () => {
        expect(
            (await getCompletionItems(getToken('select:', 0), 7, async () => []))?.map(({ label }) => label)
        ).toStrictEqual(['repo', 'file', 'content', 'symbol', 'commit'])
    })

    test('returns dynamically fetched completions', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('file:c', 0),
                    6,
                    async () =>
                        [
                            {
                                type: 'path',
                                path: 'connect.go',
                                repository: 'github.com/sourcegraph/jsonrpc2',
                            },
                        ] as SearchMatch[]
                )
            )?.map(({ label, apply }) => ({ label, apply }))
        ).toStrictEqual([{ label: 'connect.go', apply: '^connect\\.go$ ' }])
    })

    test('includes file path in insertText when completing filter value', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('f:', 0),
                    2,
                    async () =>
                        [
                            {
                                type: 'path',
                                path: 'some/path/main.go',
                                repository: 'github.com/sourcegraph/jsonrpc2',
                            },
                        ] as SearchMatch[]
                )
            )
                // Do not consider functions like has.owner or has.content.
                ?.filter(({ apply }) => typeof apply === 'string')
                .map(({ apply }) => apply)
        ).toStrictEqual(['^some/path/main\\.go$ '])
    })

    test('escapes spaces in repo value', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('repo:', 0),
                    5,
                    async () =>
                        [
                            {
                                type: 'repo',
                                repository: 'repo/with a space',
                            },
                        ] as SearchMatch[]
                )
            )
                ?.map(({ apply }) => apply)
                .filter(apply => typeof apply === 'string')
        ).toMatchInlineSnapshot(`
            [
              "^repo/with\\\\ a\\\\ space$ "
            ]
        `)
    })

    test('inserts repo: prefix for global suggestions', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('metal', 0),
                    'metal'.length,
                    async () =>
                        [
                            {
                                type: 'repo',
                                repository: 'scalameta/metals',
                            },
                        ] as SearchMatch[]
                )
            )
                ?.map(({ apply }) => apply)
                .filter(apply => typeof apply === 'string' && apply.includes('metals'))
        ).toMatchInlineSnapshot(`
            [
              "repo:^scalameta/metals$ "
            ]
        `)
    })

    test('inserts file: prefix for global suggestions', async () => {
        const query = 'repo:x local'
        const tokens = getTokens(query)
        const lastToken = tokens.at(-1)!
        expect(
            (
                await getCompletionItems(
                    lastToken,
                    query.length,
                    async () =>
                        [
                            {
                                type: 'path',
                                path: 'src/local.ts',
                                repository: 'scalameta/metals',
                            },
                        ] as SearchMatch[],
                    { tokens }
                )
            )
                ?.map(({ apply }) => apply)
                .filter(apply => typeof apply === 'string' && apply.includes('local'))
        ).toMatchInlineSnapshot(`
            [
              "file:^src/local\\\\.ts$ "
            ]
        `)
    })

    const suggestionType = (query: string): string => suggestionTypeFromTokens(getTokens(query))
    test('suggests repos for global queries', async () => {
        expect(suggestionType('sourcegraph')).toStrictEqual('repo')
    })

    test('suggests symbols for repo-scoped queries', async () => {
        expect(suggestionType('repo:sourcegraph local')).toStrictEqual('symbol')
        expect(suggestionType('r:sourcegraph local')).toStrictEqual('symbol')
    })

    test('suggests symbols for repo+file scoped queries', async () => {
        expect(suggestionType('repo:sourcegraph file:local sym')).toStrictEqual('symbol')
        expect(suggestionType('repo:sourcegraph path:local sym')).toStrictEqual('symbol')
        expect(suggestionType('repo:sourcegraph f:local sym')).toStrictEqual('symbol')
    })

    test('suggests symbols for type:symbol queries', async () => {
        expect(suggestionType('type:symbol sym')).toStrictEqual('symbol')
        expect(suggestionType('repo:sourcegraph type:symbol sym')).toStrictEqual('symbol')
        expect(suggestionType('repo:src file:local type:symbol sym')).toStrictEqual('symbol')
    })

    test('suggests files for type:path queries', async () => {
        expect(suggestionType('type:path sourcegraph')).toStrictEqual('path')
        expect(suggestionType('repo:foo type:path sourcegraph')).toStrictEqual('path')
    })
})
