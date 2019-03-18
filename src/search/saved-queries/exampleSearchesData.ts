export default [
    {
        query:
            'type:diff repo:@*refs/heads/ case:yes Copyleft|GPL|AGPL|LGPL|General.*Public.*License|Affero.*General.*Public.*License|\bMPL\b|Mozilla.*Public.*License',
        description: 'Changes mentioning a copyleft (GPL, LGPL, etc) license (all branches)',
    },
    {
        query:
            'type:diff repo:@*refs/heads/ \b(auth[^o][^r]|security\b|cve|password|secure|unsafe|perms|permissions|patch)',
        description: 'Recent security and authentication changes (all branches)',
    },
    {
        query: 'type:diff repo:@*refs/heads/ (secret|token|password)[[:alnum:]]*w*=',
        description: 'Potential secrets, tokens, or passwords (all branches)',
    },
    {
        query: 'type:diff repo:@*refs/heads/ \\/\\/\\s*(TODO|BUG|FIXME)',
        description: 'TODOs, BUGs and FIXMEs (all branches)',
    },
    {
        query: 'type:diff repo:@*refs/heads/ HACK',
        description: 'HACKs (all branches)',
    },
    {
        query: 'type:diff repo:@*refs/heads/ @deprecated|\\/\\/\\s*deprecated',
        description: 'Deprecated code (all branches)',
    },
    {
        query: 'type:diff file:(vendor|node_modules)/',
        description: 'Vendored code changes (default branch)',
    },
    {
        query: 'type:diff file:\\.(txt|md)$',
        description: '(Text/Markdown) Changes to .txt and .md files (default branch)',
    },
    {
        query: 'type:diff file:test \\bit\\(',
        description: '(TypeScript/JavaScript) Tests added / removed / modified (default branch)',
    },
    {
        query: 'type:diff file:package\\.json$',
        description: '(TypeScript/JavaScript) Dependencies added / removed / modified (default branch)',
    },
    {
        query: 'type:diff repo:@*refs/heads/ -file:extensions new\\s+[PT]?Promise|\\.then\\(|\\.catch\\(',
        description: '(TypeScript/JavaScript) non-ES6/async code changes (all branches)',
    },
    {
        query: 'type:diff repo:@*refs/heads/ this\\.setState\\(\\{.*this\\.state',
        description: '(TypeScript/JavaScript) React setState race conditions (all branches)',
    },
    {
        query: 'type:diff file:\\.s?css$',
        description: '(scss/css) Changes to .scss and .css files (default branch)',
    },
]
