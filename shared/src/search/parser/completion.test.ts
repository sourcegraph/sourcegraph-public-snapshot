import { getCompletionItems } from './completion'
import { parseSearchQuery, ParseSuccess, Sequence } from './parser'
import { NEVER, of } from 'rxjs'
import { SearchSuggestion } from '../../graphql/schema'

describe('getCompletionItems()', () => {
    test('returns static filter type completions along with dynamically fetched completions', async () => {
        expect(
            await getCompletionItems(
                're',
                (parseSearchQuery('re') as ParseSuccess<Sequence>).token,
                { column: 3 },
                () =>
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
        ).toStrictEqual({
            suggestions: [
                {
                    detail: 'Commits made after a certain date',
                    filterText: 'after',
                    insertText: 'after:',
                    kind: 22,
                    label: 'after',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '00',
                },
                {
                    detail: 'Include results from archived repositories.',
                    filterText: 'archived',
                    insertText: 'archived:',
                    kind: 22,
                    label: 'archived',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '01',
                },
                {
                    detail: 'The author of a commit',
                    filterText: 'author',
                    insertText: 'author:',
                    kind: 22,
                    label: 'author',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '02',
                },
                {
                    detail: 'Commits made before a certain date',
                    filterText: 'before',
                    insertText: 'before:',
                    kind: 22,
                    label: 'before',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '03',
                },
                {
                    detail: 'Treat the search pattern as case-sensitive.',
                    filterText: 'case',
                    insertText: 'case:',
                    kind: 22,
                    label: 'case',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '04',
                },
                {
                    detail:
                        'Explicitly overrides the search pattern. Used for explicitly delineating the search pattern to search for in case of clashes.',
                    filterText: 'content',
                    insertText: 'content:',
                    kind: 22,
                    label: 'content',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '05',
                },
                {
                    detail: 'Number of results to fetch (integer)',
                    filterText: 'count',
                    insertText: 'count:',
                    kind: 22,
                    label: 'count',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '06',
                },
                {
                    detail: 'Include only results from files matching the given regex pattern.',
                    filterText: 'file',
                    insertText: 'file:',
                    kind: 22,
                    label: 'file',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '07',
                },
                {
                    detail: 'Exclude results from files matching the given regex pattern.',
                    filterText: 'file',
                    insertText: 'file:',
                    kind: 22,
                    label: '-file',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '08',
                },
                {
                    detail: 'Include results from forked repositories.',
                    filterText: 'fork',
                    insertText: 'fork:',
                    kind: 22,
                    label: 'fork',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '09',
                },
                {
                    detail: 'Include only results from the given language',
                    filterText: 'lang',
                    insertText: 'lang:',
                    kind: 22,
                    label: 'lang',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '010',
                },
                {
                    detail: 'Exclude results from the given language',
                    filterText: 'lang',
                    insertText: 'lang:',
                    kind: 22,
                    label: '-lang',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '011',
                },
                {
                    detail: 'Commits with messages matching a certain string',
                    filterText: 'message',
                    insertText: 'message:',
                    kind: 22,
                    label: 'message',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '012',
                },
                {
                    detail: 'The pattern type (regexp, literal, structural) in use',
                    filterText: 'patterntype',
                    insertText: 'patterntype:',
                    kind: 22,
                    label: 'patterntype',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '013',
                },
                {
                    detail: 'Include only results from repositories matching the given regex pattern.',
                    filterText: 'repo',
                    insertText: 'repo:',
                    kind: 22,
                    label: 'repo',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '014',
                },
                {
                    detail: 'Exclude results from repositories matching the given regex pattern.',
                    filterText: 'repo',
                    insertText: 'repo:',
                    kind: 22,
                    label: '-repo',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '015',
                },
                {
                    detail: 'group-name (include results from the named group)',
                    filterText: 'repogroup',
                    insertText: 'repogroup:',
                    kind: 22,
                    label: 'repogroup',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '016',
                },
                {
                    detail: '"string specifying time frame" (filter out stale repositories without recent commits)',
                    filterText: 'repohascommitafter',
                    insertText: 'repohascommitafter:',
                    kind: 22,
                    label: 'repohascommitafter',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '017',
                },
                {
                    detail: 'Include only results from repos that contain a matching file',
                    filterText: 'repohasfile',
                    insertText: 'repohasfile:',
                    kind: 22,
                    label: 'repohasfile',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '018',
                },
                {
                    detail: 'Exclude results from repos that contain a matching file',
                    filterText: 'repohasfile',
                    insertText: 'repohasfile:',
                    kind: 22,
                    label: '-repohasfile',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '019',
                },
                {
                    detail: 'Duration before timeout',
                    filterText: 'timeout',
                    insertText: 'timeout:',
                    kind: 22,
                    label: 'timeout',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '020',
                },
                {
                    detail: 'Limit results to the specified type.',
                    filterText: 'type',
                    insertText: 'type:',
                    kind: 22,
                    label: 'type',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '021',
                },
                {
                    filterText: 'github.com/sourcegraph/jsonrpc2',
                    insertText: '^github\\.com/sourcegraph/jsonrpc2$ ',
                    kind: 19,
                    label: 'github.com/sourcegraph/jsonrpc2',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '1',
                },
                {
                    detail: 'Variable - github.com/sourcegraph/jsonrpc2',
                    filterText: 'RepoRoutes',
                    insertText: 'RepoRoutes',
                    kind: 4,
                    label: 'RepoRoutes',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                    sortText: '1',
                },
            ],
        })
    })

    test('returns suggestions for an empty query', async () => {
        expect(
            (
                await getCompletionItems(
                    '',
                    (parseSearchQuery('') as ParseSuccess<Sequence>).token,
                    { column: 1 },
                    () => NEVER
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
                    'a ',
                    (parseSearchQuery('a ') as ParseSuccess<Sequence>).token,
                    { column: 3 },
                    () =>
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
                    'rE',
                    (parseSearchQuery('rE') as ParseSuccess<Sequence>).token,
                    { column: 3 },
                    () => of([])
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
                    'case:y',
                    (parseSearchQuery('case:y') as ParseSuccess<Sequence>).token,
                    { column: 7 },
                    () => NEVER
                )
            )?.suggestions.map(({ label }) => label)
        ).toStrictEqual(['yes', 'no'])
    })

    test('returns completions for filters with static suggestions', async () => {
        expect(
            (
                await getCompletionItems(
                    'lang:',
                    (parseSearchQuery('lang:') as ParseSuccess<Sequence>).token,
                    {
                        column: 6,
                    },
                    () => of([])
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
                    'a',
                    (parseSearchQuery('file:c') as ParseSuccess<Sequence>).token,
                    { column: 7 },
                    () =>
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
