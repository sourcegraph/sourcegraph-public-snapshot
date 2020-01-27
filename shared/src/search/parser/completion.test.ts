import { getCompletionItems } from './completion'
import { parseSearchQuery, ParseSuccess, Sequence } from './parser'
import { NEVER, of } from 'rxjs'
import { IFile } from '../../graphql/schema'

describe('getCompletionItems()', () => {
    test('returns static filter type completions', async () => {
        expect(
            await getCompletionItems(
                'a',
                (parseSearchQuery('re') as ParseSuccess<Sequence>).token,
                { column: 3 },
                () => NEVER
            )
        ).toStrictEqual({
            suggestions: [
                {
                    detail: 'Include only results from repositories matching the given regex pattern.',
                    filterText: 'repo',
                    insertText: 'repo:',
                    kind: 17,
                    label: 'repo',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
                {
                    detail: 'group-name (include results from the named group)',
                    filterText: 'repogroup',
                    insertText: 'repogroup:',
                    kind: 17,
                    label: 'repogroup',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
                {
                    detail: 'regex-pattern (include results from repos that contain a matching file)',
                    filterText: 'repohasfile',
                    insertText: 'repohasfile:',
                    kind: 17,
                    label: 'repohasfile',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
                {
                    detail: '"string specifying time frame" (filter out stale repositories without recent commits)',
                    filterText: 'repohascommitafter',
                    insertText: 'repohascommitafter:',
                    kind: 17,
                    label: 'repohascommitafter',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
            ],
        })
    })

    test('returns static filter type completions for case-insensitive query', async () => {
        expect(
            await getCompletionItems(
                'a',
                (parseSearchQuery('rE') as ParseSuccess<Sequence>).token,
                { column: 3 },
                () => NEVER
            )
        ).toStrictEqual({
            suggestions: [
                {
                    detail: 'Include only results from repositories matching the given regex pattern.',
                    filterText: 'repo',
                    insertText: 'repo:',
                    kind: 17,
                    label: 'repo',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
                {
                    detail: 'group-name (include results from the named group)',
                    filterText: 'repogroup',
                    insertText: 'repogroup:',
                    kind: 17,
                    label: 'repogroup',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
                {
                    detail: 'regex-pattern (include results from repos that contain a matching file)',
                    filterText: 'repohasfile',
                    insertText: 'repohasfile:',
                    kind: 17,
                    label: 'repohasfile',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
                {
                    detail: '"string specifying time frame" (filter out stale repositories without recent commits)',
                    filterText: 'repohascommitafter',
                    insertText: 'repohascommitafter:',
                    kind: 17,
                    label: 'repohascommitafter',
                    range: {
                        endColumn: 3,
                        endLineNumber: 1,
                        startColumn: 1,
                        startLineNumber: 1,
                    },
                },
            ],
        })
    })

    test('returns completions for filters with discrete values', async () => {
        expect(
            await getCompletionItems(
                'a',
                (parseSearchQuery('case:y') as ParseSuccess<Sequence>).token,
                { column: 7 },
                () => NEVER
            )
        ).toStrictEqual({
            suggestions: [
                {
                    filterText: 'yes',
                    insertText: 'yes ',
                    kind: 18,
                    label: 'yes',
                    range: {
                        endColumn: 7,
                        endLineNumber: 1,
                        startColumn: 6,
                        startLineNumber: 1,
                    },
                },
                {
                    filterText: 'no',
                    insertText: 'no ',
                    kind: 18,
                    label: 'no',
                    range: {
                        endColumn: 7,
                        endLineNumber: 1,
                        startColumn: 6,
                        startLineNumber: 1,
                    },
                },
            ],
        })
    })

    test('returns dynamically fetched completions', async () => {
        expect(
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
                            isDirectory: false,
                            url: 'b',
                            repository: {
                                name: 'r',
                            },
                        } as IFile,
                    ])
            )
        ).toStrictEqual({
            suggestions: [
                {
                    detail: 'connect.go - r',
                    filterText: 'file:connect.go',
                    insertText: '^connect\\.go$ ',
                    kind: 18,
                    label: 'connect.go',
                    range: {
                        endColumn: 7,
                        endLineNumber: 1,
                        startColumn: 6,
                        startLineNumber: 1,
                    },
                },
            ],
        })
    })
})
