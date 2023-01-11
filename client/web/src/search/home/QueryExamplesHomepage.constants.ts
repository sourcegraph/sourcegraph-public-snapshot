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
            title: 'Scope search to specific repos',
            queryExamples: [
                { id: 'org-repos', query: 'repo:sourcegraph/.*' },
                { id: 'single-repo', query: 'repo:facebook/react' },
            ],
        },
        {
            title: 'Jump into code navigation',
            queryExamples: [
                { id: 'file-filter', query: 'file:README.md' },
                { id: 'type-symbol', query: 'type:symbol SymbolName' },
            ],
        },
        {
            title: 'Get logical',
            queryExamples: [
                { id: 'not-operator', query: 'lang:go NOT file:main.go' },
                { id: 'or-operator', query: 'lang:javascript OR lang:typescript' },
                { id: 'and-operator', query: 'hello AND world' },
            ],
        },
    ],
    [
        {
            title: 'Find content or patterns',
            queryExamples: [
                { id: 'exact-matches', query: 'some exact error message', helperText: 'No quotes needed' },
                { id: 'regex-pattern', query: '/regex.*pattern/' },
            ],
        },
        {
            title: 'Explore code history',
            queryExamples: [
                { id: 'type-diff-author', query: 'type:diff author:torvalds' },
                { id: 'type-commit-message', query: 'type:commit some message' },
            ],
        },
        {
            title: 'Get advanced',
            queryExamples: [
                { id: 'repo-has-description', query: 'repo:has.description(scientific computing)' },
                { id: 'commit-search', query: 'repo:has.commit.after(june 25 2017)' },
            ],
        },
    ],
]
