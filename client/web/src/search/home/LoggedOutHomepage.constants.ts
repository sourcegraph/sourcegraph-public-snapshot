import { DynamicWebFont } from './DynamicWebFonts'

export interface SearchExample {
    label: string
    trackEventName: string
    query: string
    to: string
}

export const exampleQueries: SearchExample[] = [
    {
        label: 'Search all of your repos, without escaping or regex',
        trackEventName: 'HomepageExampleRepoClicked',
        query: 'repo:sourcegraph/.* Sprintf("%d -file:tests',
        to: '/search?q=context:global+repo:sourcegraph/.*+Sprintf%28%22%25d+-file:tests&patternType=literal&case=yes',
    },
    {
        label: 'Search and review commits faster than git log and grep',
        trackEventName: 'HomepageExampleDiffClicked',
        query: 'type:diff before:"last week" TODO',
        to:
            '/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+type:diff+after:"last+week"+select:commit.diff.added+TODO&patternType=literal&case=yes',
    },
    {
        label: 'Quickly filter by language and other key attributes',
        trackEventName: 'HomepageExampleFiltersClicked',
        query: 'repo:sourcegraph lang:go or lang:Typescript',
        to:
            '/search?q=context:global+repo:sourcegraph/*+-f:tests+%28lang:TypeScript+or+lang:go%29+Config%28%29&patternType=literal&case=yes',
    },
]

export const exampleNotebooks: SearchExample[] = [
    {
        label: 'Find and reference code across all of your repositories',
        trackEventName: 'HomepageNotebookRepoClicked',
        query: 'repo:sourcegraph/.* Config()',
        to: '/github.com/sourcegraph/notebooks/-/blob/onboarding/find-code-across-all-of-your-repositories.snb.md',
    },
    {
        label: 'Search and review commits and their code faster than git log and grep ',
        trackEventName: 'HomepageNotebookDiffClicked',
        query: 'type:diff before:"last week" TODO',
        to: '/github.com/sourcegraph/notebooks/-/blob/onboarding/search-and-review-commits.snb.md',
    },
    {
        label: 'Quickly filter by file path, language and other elements of code',
        trackEventName: 'HomepageNotebookFiltersClicked',
        query: 'repo:sourcegraph/.* lang:go -f:tests',
        to:
            '/github.com/sourcegraph/notebooks/-/blob/onboarding/filter-by-file-language-and-other-elements-of-code.snb.md',
    },
]

/**
 * Source Sans Pro fonts: https://fonts.google.com/specimen/Source+Sans+Pro
 * Two families required for the `LoggedOutHomepage` UI: Regular 400 and Bold 700.
 *
 * Assets are downloaded from the Google generated link. Only latin glyphs are included.
 * https://fonts.googleapis.com/css2?family=Source+Sans+Pro:wght@400;700&display=swap
 */
export const fonts: DynamicWebFont[] = [
    {
        family: 'Source Sans Pro',
        weight: '400',
        source: "url(/.assets/SourceSansPro-Regular.woff2) format('woff2')",
        style: 'normal',
        unicodeRange:
            'U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD',
    },
    {
        family: 'Source Sans Pro',
        weight: '700',
        source: "url(/.assets/SourceSansPro-Bold.woff2) format('woff2')",
        style: 'normal',
        unicodeRange:
            'U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+2000-206F, U+2074, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD',
    },
]
