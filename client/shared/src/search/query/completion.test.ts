import { SymbolKind } from '../../graphql-operations'
import { isSearchMatchOfType, SearchMatch } from '../stream'

import { FetchSuggestions, getCompletionItems } from './completion'
import { POPULAR_LANGUAGES } from './languageFilter'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

const getToken = (query: string, tokenIndex: number): Token => toSuccess(scanSearchQuery(query))[tokenIndex]

const createFetcher = (matches: SearchMatch[]): FetchSuggestions => (_token, type) =>
    Promise.resolve(matches.filter(isSearchMatchOfType(type)))

// Using async as a short way to create functions that return promises
/* eslint-disable @typescript-eslint/require-await */
describe('getCompletionItems()', () => {
    test('returns only static filter type completions when the token matches a known filter', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('re', 0),
                    { column: 3 },
                    createFetcher([
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            type: 'symbol',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                            path: '',
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
                    ]),
                    {}
                )
            )?.suggestions.map(({ label }) => label)
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
        ])
    })

    test("returns static filter type completions along with dynamically fetched completions when the token doesn't match a filter", async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('reposi', 0),
                    { column: 7 },
                    createFetcher([
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            type: 'symbol',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                            path: '',
                            symbols: [
                                {
                                    kind: SymbolKind.VARIABLE,
                                    name: 'RepoRoutes',
                                    containerName: '',
                                    url: '',
                                    line: 1,
                                },
                            ],
                        },
                    ]),
                    {}
                )
            )?.suggestions.map(({ label }) => label)
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
            'RepoRoutes',
        ])
    })

    test('returns suggestions for an empty query', async () => {
        expect(
            (await getCompletionItems(getToken('', 0), { column: 1 }, createFetcher([]), {}))?.suggestions.map(
                ({ label }) => label
            )
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
        ])
    })

    test('returns suggestions on whitespace', async () => {
        expect(
            (await getCompletionItems(getToken('a ', 1), { column: 3 }, createFetcher([]), {}))?.suggestions.map(
                ({ label }) => label
            )
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
        ])
    })

    test('returns static filter type completions for case-insensitive query', async () => {
        expect(
            (await getCompletionItems(getToken('rE', 0), { column: 3 }, createFetcher([]), {}))?.suggestions.map(
                ({ label }) => label
            )
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
        ])
    })

    test('returns completions for filters with discrete values', async () => {
        expect(
            (await getCompletionItems(getToken('case:y', 0), { column: 7 }, createFetcher([]), {}))?.suggestions.map(
                ({ label }) => label
            )
        ).toStrictEqual(['yes', 'no'])
    })

    test('returns completions for filters with static suggestions', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('lang:', 0),
                    {
                        column: 6,
                    },
                    createFetcher([]),
                    {}
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual(POPULAR_LANGUAGES)
    })

    test('returns completions in order of discrete value definition, not alphabetically', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('select:', 0),
                    {
                        column: 8,
                    },
                    createFetcher([]),
                    {}
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual(['repo', 'file', 'content', 'symbol', 'commit'])
    })

    test('returns dynamically fetched completions', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('file:c', 0),
                    { column: 7 },
                    createFetcher([
                        {
                            type: 'path',
                            path: 'connect.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ]),
                    {}
                )
            )?.suggestions.map(({ label, insertText }) => ({ label, insertText }))
        ).toStrictEqual([{ label: 'connect.go', insertText: '^connect\\.go$ ' }])
    })

    test('sets current filter value as filterText', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('f:^jsonrpc', 0),
                    { column: 11 },
                    createFetcher([
                        {
                            type: 'path',
                            path: 'jsonrpc2.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ]),
                    {}
                )
            )?.suggestions.map(({ filterText }) => filterText)
        ).toStrictEqual(['^jsonrpc'])
    })

    test('includes file path in insertText when completing filter value', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('f:', 0),
                    { column: 2 },
                    createFetcher([
                        {
                            type: 'path',
                            path: 'some/path/main.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ]),
                    {}
                )
            )?.suggestions.map(({ insertText }) => insertText)
        ).toStrictEqual(['^some/path/main\\.go$ '])
    })

    test('escapes spaces in repo value', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('repo:', 0),
                    { column: 5 },
                    createFetcher([
                        {
                            type: 'repo',
                            repository: 'repo/with a space',
                        },
                    ]),

                    {}
                )
            )?.suggestions.map(({ insertText }) => insertText)
        ).toMatchInlineSnapshot(`
            [
              "has.path(\${1:CHANGELOG}) ",
              "has.content(\${1:TODO}) ",
              "has.file(path:\${1:CHANGELOG} content:\${2:fix}) ",
              "has.commit.after(\${1:1 month ago}) ",
              "has.description(\${1}) ",
              "^repo/with\\\\ a\\\\ space$ "
            ]
        `)
    })

    test('Sourcegraph.com GH repo completions', async () => {
        expect(
            (
                await getCompletionItems(getToken('repo:', 0), { column: 5 }, createFetcher([]), {
                    isSourcegraphDotCom: true,
                })
            )?.suggestions.map(({ insertText }) => insertText)
        ).toMatchInlineSnapshot(`
            [
              "^github\\\\.com/\${1:ORGANIZATION}/.* ",
              "^github\\\\.com/\${1:ORGANIZATION}/\${2:REPO-NAME}$ ",
              "\${1:STRING} ",
              "has.path(\${1:CHANGELOG}) ",
              "has.content(\${1:TODO}) ",
              "has.file(path:\${1:CHANGELOG} content:\${2:fix}) ",
              "has.commit.after(\${1:1 month ago}) ",
              "has.description(\${1}) "
            ]
        `)
    })
})
