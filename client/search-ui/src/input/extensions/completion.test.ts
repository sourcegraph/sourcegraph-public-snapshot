import { Completion } from '@codemirror/autocomplete'

import { SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { POPULAR_LANGUAGES } from '@sourcegraph/shared/src/search/query/languageFilter'
import { ScanResult, scanSearchQuery, ScanSuccess } from '@sourcegraph/shared/src/search/query/scanner'
import { Token } from '@sourcegraph/shared/src/search/query/token'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { createDefaultSuggestionSources } from './completion'

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
    options?: { globbing?: boolean; isSourcegraphDotCom?: boolean }
) {
    const sources = createDefaultSuggestionSources({
        globbing: false,
        isSourcegraphDotCom: false,
        fetchSuggestions,
        ...options,
    })

    const results = await Promise.all(sources.map(source => source({ position, onAbort: () => {} }, [token], token)))
    const allOptions: Completion[] = []
    for (const result of results) {
        if (result) {
            allOptions.push(...result.options)
        }
    }
    return allOptions
}

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

const getToken = (query: string, tokenIndex: number): Token => toSuccess(scanSearchQuery(query))[tokenIndex]

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
                            },
                        ],
                    },
                ])
            )?.map(({ label }) => label)
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
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
            'RepoRoutes',
        ])
    })

    test('returns suggestions for an empty query', async () => {
        expect(
            (await getCompletionItems(getToken('', 0), 0, async () => []))?.map(({ label }) => label)
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
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
        ])
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
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
        ])
    })

    test('returns static filter type completions for case-insensitive query', async () => {
        expect(
            (await getCompletionItems(getToken('rE', 0), 2, async () => []))?.map(({ label }) => label)
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
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
        ])
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
                        ] as SearchMatch[],
                    {}
                )
            )?.map(({ label, apply }) => ({ label, apply }))
        ).toStrictEqual([{ label: 'connect.go', apply: '^connect\\.go$ ' }])
    })

    test('inserts valid suggestion when completing repo:deps predicate', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('repo:deps(sourcegraph', 0),
                    20,
                    async () =>
                        [
                            {
                                type: 'repo',
                                repository: 'github.com/sourcegraph/jsonrpc2.go',
                            },
                        ] as SearchMatch[]
                )
            )?.map(({ apply }) => apply)
        ).toStrictEqual(['deps(^github\\.com/sourcegraph/jsonrpc2\\.go$) '])
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
            )?.map(({ apply }) => apply)
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
})
