import * as Monaco from 'monaco-editor'
import { NEVER, of } from 'rxjs'

import { SymbolKind } from '../../graphql-operations'
import { SearchMatch } from '../stream'

import { getCompletionItems, repositoryCompletionItemKind } from './completion'
import { POPULAR_LANGUAGES } from './languageFilter'
import { scanSearchQuery, ScanSuccess, ScanResult } from './scanner'
import { Token } from './token'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value, null, 2),
    test: () => true,
})

const toSuccess = (result: ScanResult<Token[]>): Token[] => (result as ScanSuccess<Token[]>).term

const getToken = (query: string, tokenIndex: number): Token => toSuccess(scanSearchQuery(query))[tokenIndex]

describe('getCompletionItems()', () => {
    test('returns only static filter type completions when the token matches a known filter', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('re', 0),
                    { column: 3 },
                    of([
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            type: 'symbol',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                            symbols: [
                                {
                                    kind: SymbolKind.VARIABLE,
                                    name: 'RepoRoutes',
                                },
                            ],
                        },
                    ] as SearchMatch[]),
                    false
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

    test("returns static filter type completions along with dynamically fetched completions when the token doesn't match a filter", async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('reposi', 0),
                    { column: 7 },
                    of([
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            type: 'symbol',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                            symbols: [
                                {
                                    kind: SymbolKind.VARIABLE,
                                    name: 'RepoRoutes',
                                },
                            ],
                        },
                    ] as SearchMatch[]),
                    false
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
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
            'github.com/sourcegraph/jsonrpc2',
            'RepoRoutes',
        ])
    })

    test('returns suggestions for an empty query', async () => {
        expect(
            (await getCompletionItems(getToken('', 0), { column: 1 }, NEVER, false))?.suggestions.map(
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
            (
                await getCompletionItems(
                    getToken('a ', 1),
                    { column: 3 },
                    of([
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ] as SearchMatch[]),
                    false
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
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'rev',
            'select',
            'timeout',
            'type',
            'visibility',
            'github.com/sourcegraph/jsonrpc2',
        ])
    })

    test('returns static filter type completions for case-insensitive query', async () => {
        expect(
            (await getCompletionItems(getToken('rE', 0), { column: 3 }, of([]), false))?.suggestions.map(
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
            (await getCompletionItems(getToken('case:y', 0), { column: 7 }, NEVER, false))?.suggestions.map(
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
                    of([]),
                    false
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
                    of([]),
                    false
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
                    of([
                        {
                            type: 'path',
                            path: 'connect.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ] as SearchMatch[]),
                    false
                )
            )?.suggestions.map(({ label, insertText }) => ({ label, insertText }))
        ).toStrictEqual([{ label: 'connect.go', insertText: '^connect\\.go$ ' }])
    })

    test('inserts valid filters when selecting a file or repository suggestion', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('jsonrpc', 0),
                    { column: 8 },
                    of([
                        {
                            type: 'path',
                            path: 'jsonrpc2.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2.go',
                        },
                    ] as SearchMatch[]),
                    false
                )
            )?.suggestions
                .filter(
                    ({ kind }) =>
                        kind === Monaco.languages.CompletionItemKind.File || kind === repositoryCompletionItemKind
                )
                .map(({ insertText }) => insertText)
        ).toStrictEqual(['file:^jsonrpc2\\.go$ ', 'repo:^github\\.com/sourcegraph/jsonrpc2\\.go$ '])
    })

    test('inserts valid suggestion when completing repo:deps predicate', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('repo:deps(sourcegraph', 0),
                    { column: 21 },
                    of([
                        {
                            type: 'repo',
                            repository: 'github.com/sourcegraph/jsonrpc2.go',
                        },
                    ] as SearchMatch[]),
                    false
                )
            )?.suggestions
                .filter(({ kind }) => kind === repositoryCompletionItemKind)
                .map(({ insertText }) => insertText)
        ).toStrictEqual(['deps(^github\\.com/sourcegraph/jsonrpc2\\.go$) '])
    })

    test('sets current filter value as filterText', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('f:^jsonrpc', 0),
                    { column: 11 },
                    of([
                        {
                            type: 'path',
                            path: 'jsonrpc2.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ] as SearchMatch[]),
                    false
                )
            )?.suggestions.map(({ filterText }) => filterText)
        ).toStrictEqual(['^jsonrpc'])
    })

    test('includes file path in insertText with fuzzy completions', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('main.go', 0),
                    { column: 7 },
                    of([
                        {
                            type: 'path',
                            path: 'some/path/main.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ] as SearchMatch[]),
                    false
                )
            )?.suggestions
                .filter(({ insertText }) => insertText.includes('some/path'))
                .map(({ insertText }) => insertText)
        ).toStrictEqual(['file:^some/path/main\\.go$ '])
    })

    test('includes file path in insertText when completing filter value', async () => {
        expect(
            (
                await getCompletionItems(
                    getToken('f:', 0),
                    { column: 2 },
                    of([
                        {
                            type: 'path',
                            path: 'some/path/main.go',
                            repository: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ] as SearchMatch[]),
                    false
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
                    of([
                        {
                            type: 'repo',
                            repository: 'repo/with a space',
                        },
                    ] as SearchMatch[]),
                    false
                )
            )?.suggestions.map(({ insertText }) => insertText)
        ).toMatchInlineSnapshot(`
            [
              "contains.file(\${1:CHANGELOG}) ",
              "contains.content(\${1:TODO}) ",
              "contains(file:\${1:CHANGELOG} content:\${2:fix}) ",
              "contains.commit.after(\${1:1 month ago}) ",
              "deps(\${1}) ",
              "dependencies(\${1}) ",
              "^repo/with\\\\ a\\\\ space$ "
            ]
        `)
    })

    test('Sourcegraph.com GH repo completions', async () => {
        expect(
            (await getCompletionItems(getToken('repo:', 0), { column: 5 }, of([]), false, true))?.suggestions.map(
                ({ insertText }) => insertText
            )
        ).toMatchInlineSnapshot(`
            [
              "^github\\\\.com/\${1:ORGANIZATION}/.* ",
              "^github\\\\.com/\${1:ORGANIZATION}/\${2:REPO-NAME}$ ",
              "\${1:STRING} ",
              "contains.file(\${1:CHANGELOG}) ",
              "contains.content(\${1:TODO}) ",
              "contains(file:\${1:CHANGELOG} content:\${2:fix}) ",
              "contains.commit.after(\${1:1 month ago}) ",
              "deps(\${1}) ",
              "dependencies(\${1}) "
            ]
        `)
    })
})
