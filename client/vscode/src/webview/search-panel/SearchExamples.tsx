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
