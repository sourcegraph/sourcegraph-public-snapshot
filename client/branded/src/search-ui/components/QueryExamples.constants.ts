/** Static query examples */

export const exampleQueryColumns = [
    [
        {
            title: 'Find to-dos on a specific repository',
            queryExamples: [
                {
                    id: 'find-todos',
                    query: 'repo:facebook/react content:TODO',
                    slug: '?q=context%3Aglobal+repo%3Afacebook%2Freact+content%3ATODO',
                },
            ],
        },
        {
            title: 'Error boundaries in React',
            queryExamples: [
                {
                    id: 'error-boundaries',
                    query: 'static getDerivedStateFromError(',
                    slug: '?q=context%3Aglobal+static+getDerivedStateFromError%28',
                },
            ],
        },
    ],
    [
        {
            title: 'Discover how developers are using hooks',
            queryExamples: [
                {
                    id: 'hooks',
                    query: 'useState AND useRef lang:javascript',
                    slug: '?q=context:global+useState+AND+useRef+lang:javascript',
                },
            ],
        },
        {
            title: "Type check, find what's passed to propTypes",
            queryExamples: [
                {
                    id: 'prop-types',
                    query: '.propTypes = {...} patterntype:structural',
                    slug: '?q=context:global+.propTypes+%3D+%7B...%7D',
                },
            ],
        },
    ],
]

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
