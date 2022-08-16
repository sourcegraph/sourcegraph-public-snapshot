export interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
    group?: string
}

// Mock data for bar chart, will be removed and replace with
// actual data in https://github.com/sourcegraph/sourcegraph/issues/39956
export const LANGUAGE_USAGE_DATA: LanguageUsageDatum[] = [
    {
        name: 'github/sourcegraph/Julia',
        value: 1000,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Erlang',
        value: 700,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'github/sourcegraph/sourcegraph/SQL',
        value: 550,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Cobol',
        value: 500,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'github/sourcegraph/sourcegraph/JavaScript',
        value: 422,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'github/sourcegraph/sourcegraph/CSS',
        value: 273,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        name: 'github/sourcegraph/sourcegraph/HTML',
        value: 129,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        name: 'github/sourcegraph/sourcegraph/С++',
        value: 110,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/TypeScript',
        value: 95,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Elm',
        value: 84,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Rust',
        value: 60,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Go',
        value: 45,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Markdown',
        value: 35,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/Zig',
        value: 20,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'github/sourcegraph/sourcegraph/XML',
        value: 5,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]
