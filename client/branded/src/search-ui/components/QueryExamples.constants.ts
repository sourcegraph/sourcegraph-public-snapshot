/** Static query examples */

export const basicSyntaxColumns = [
    [
        {
            title: 'Search in files',
            queryExamples: [
                { id: 'exact-matches', query: 'some error message', helperText: 'No quotes needed' },
                { id: 'regex-pattern', query: '/open(File|Dir)/' },
                { id: 'file', query: 'file:README foo' },
            ],
        },
        {
            title: 'Search in commit diffs',
            queryExamples: [{ id: 'type-diff-author', query: 'repo:sourcegraph$ type:diff fix' }],
        },
    ],
    [
        {
            title: 'Filter by...',
            queryExamples: [
                { id: 'single-repo', query: 'repo:sourcegraph/sourcegraph' },
                { id: 'org-repos', query: 'repo:facebook/react' },
                { id: 'repo-has-description', query: 'repo:has.description(GPT)' },
                { id: 'lang-filter', query: 'lang:typescript' },
            ],
        },
    ],
]
