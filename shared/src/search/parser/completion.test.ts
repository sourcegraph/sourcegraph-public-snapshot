import { getCompletionItems } from './completion'
import { parseSearchQuery, ParseSuccess, Sequence } from './parser'
import { NEVER, of } from 'rxjs'
import { SearchSuggestion } from '../../graphql/schema'

describe('getCompletionItems()', () => {
    test('returns only static filter type completions when the token matches a known filter', async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('re') as ParseSuccess<Sequence>).token,
                    { column: 3 },
                    of([
                        {
                            __typename: 'Repository',
                            name: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            __typename: 'Symbol',
                            name: 'RepoRoutes',
                            kind: 'VARIABLE',
                            location: {
                                resource: {
                                    repository: {
                                        name: 'github.com/sourcegraph/jsonrpc2',
                                    },
                                },
                            },
                        },
                    ] as SearchSuggestion[])
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual([
            'after',
            'archived',
            'author',
            'before',
            'case',
            'content',
            'count',
            'file',
            '-file',
            'fork',
            'index',
            'lang',
            '-lang',
            'message',
            'patterntype',
            'repo',
            '-repo',
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'timeout',
            'type',
        ])
    })

    test("returns static filter type completions along with dynamically fetched completions when the token doesn't match a filter", async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('reposi') as ParseSuccess<Sequence>).token,
                    { column: 7 },
                    of([
                        {
                            __typename: 'Repository',
                            name: 'github.com/sourcegraph/jsonrpc2',
                        },
                        {
                            __typename: 'Symbol',
                            name: 'RepoRoutes',
                            kind: 'VARIABLE',
                            location: {
                                resource: {
                                    repository: {
                                        name: 'github.com/sourcegraph/jsonrpc2',
                                    },
                                },
                            },
                        },
                    ] as SearchSuggestion[])
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual([
            'after',
            'archived',
            'author',
            'before',
            'case',
            'content',
            'count',
            'file',
            '-file',
            'fork',
            'index',
            'lang',
            '-lang',
            'message',
            'patterntype',
            'repo',
            '-repo',
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'timeout',
            'type',
            'github.com/sourcegraph/jsonrpc2',
            'RepoRoutes',
        ])
    })

    test('returns suggestions for an empty query', async () => {
        expect(
            (
                await getCompletionItems((parseSearchQuery('') as ParseSuccess<Sequence>).token, { column: 1 }, NEVER)
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual([
            'after',
            'archived',
            'author',
            'before',
            'case',
            'content',
            'count',
            'file',
            '-file',
            'fork',
            'index',
            'lang',
            '-lang',
            'message',
            'patterntype',
            'repo',
            '-repo',
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'timeout',
            'type',
        ])
    })

    test('returns suggestions on whitespace', async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('a ') as ParseSuccess<Sequence>).token,
                    { column: 3 },
                    of([
                        {
                            __typename: 'Repository',
                            name: 'github.com/sourcegraph/jsonrpc2',
                        },
                    ] as SearchSuggestion[])
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual([
            'after',
            'archived',
            'author',
            'before',
            'case',
            'content',
            'count',
            'file',
            '-file',
            'fork',
            'index',
            'lang',
            '-lang',
            'message',
            'patterntype',
            'repo',
            '-repo',
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'timeout',
            'type',
            'github.com/sourcegraph/jsonrpc2',
        ])
    })

    test('returns static filter type completions for case-insensitive query', async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('rE') as ParseSuccess<Sequence>).token,
                    { column: 3 },
                    of([])
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual([
            'after',
            'archived',
            'author',
            'before',
            'case',
            'content',
            'count',
            'file',
            '-file',
            'fork',
            'index',
            'lang',
            '-lang',
            'message',
            'patterntype',
            'repo',
            '-repo',
            'repogroup',
            'repohascommitafter',
            'repohasfile',
            '-repohasfile',
            'timeout',
            'type',
        ])
    })

    test('returns completions for filters with discrete values', async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('case:y') as ParseSuccess<Sequence>).token,
                    { column: 7 },
                    NEVER
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual(['yes', 'no'])
    })

    test('returns completions for filters with static suggestions', async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('lang:') as ParseSuccess<Sequence>).token,
                    {
                        column: 6,
                    },
                    of([])
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual([
            'c',
            'cpp',
            'csharp',
            'css',
            'go',
            'graphql',
            'haskell',
            'html',
            'java',
            'javascript',
            'json',
            'lua',
            'markdown',
            'php',
            'powershell',
            'python',
            'r',
            'ruby',
            'sass',
            'swift',
            'typescript',
        ])
    })

    test('returns dynamically fetched completions', async () => {
        expect(
            (
                await getCompletionItems(
                    (parseSearchQuery('file:c') as ParseSuccess<Sequence>).token,
                    { column: 7 },
                    of([
                        {
                            __typename: 'File',
                            path: 'connect.go',
                            name: 'connect.go',
                            repository: {
                                name: 'github.com/sourcegraph/jsonrpc2',
                            },
                        },
                    ] as SearchSuggestion[])
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual(['connect.go'])
    })
})
