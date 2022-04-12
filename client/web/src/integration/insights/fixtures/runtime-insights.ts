import { BulkSearchCommits, BulkSearchFields, SearchResultsStatsResult } from '../../../graphql-operations'

/**
 * All just in time insights on the fronted since we don't have pre-calculated data for them in
 * the code insights database tables work through our Search API gql handlers and have two steps
 * pipeline.
 *
 * 1. FE logic calculates list of dates for possible commits in insight's repositories field -
 * BulkSearchCommits mock
 *
 * 2. Then we query insight query for each commit that we got on the first step and process search
 * matches - BulkSearchCommits
 *
 * Note that code stats insight works in one-step pipeline since we don't history quries there and
 * always show the latest data for code stat insight repository.
 */

/**
 * Metadata for just in time (FE) insight commit search. This fixture provides commits metadata
 * - the first step in just in time insight processing
 */
export const STORYBOOK_GROWTH_INSIGHT_COMMITS_FIXTURE: Record<string, BulkSearchCommits> = {
    search0: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '0b81b624c24f5fc4d53fd10651eb84d67072e74e',
                        committer: {
                            date: '2009-09-04T05:00:34Z',
                        },
                    },
                },
            ],
        },
    },
    search1: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '03956c5dde9dcb22a53e7d3f259a0e98dd50704b',
                        committer: {
                            date: '2011-09-05T19:10:27Z',
                        },
                    },
                },
            ],
        },
    },
    search2: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '948698c42a2a05b1ebebf3cda5945b0764426f57',
                        committer: {
                            date: '2013-09-04T22:25:47Z',
                        },
                    },
                },
            ],
        },
    },
    search3: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: 'db37595dc3a4ceb7953dd89b2122a974ff70b311',
                        committer: {
                            date: '2015-09-05T18:20:08Z',
                        },
                    },
                },
            ],
        },
    },
    search4: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '4472874fb5fddb44d402f4fdf064149b68dce68c',
                        committer: {
                            date: '2017-09-05T16:16:25Z',
                        },
                    },
                },
            ],
        },
    },
    search5: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '6f34ec9377091a1665d554a54bc58812813c7c1a',
                        committer: {
                            date: '2019-07-09T19:59:43Z',
                        },
                    },
                },
            ],
        },
    },
    search6: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '6f34ec9377091a1665d554a54bc58812813c7c1a',
                        committer: {
                            date: '2019-07-09T19:59:43Z',
                        },
                    },
                },
            ],
        },
    },
}

/**
 * This fixture provides search API matches data - the second step in just in time processing
 * data.
 */
export const STORYBOOK_GROWTH_INSIGHT_MATCH_DATA_FIXTURE: Record<string, BulkSearchFields> = {
    search0: { results: { matchCount: 10245 } },
    search1: { results: { matchCount: 16502 } },
    search2: { results: { matchCount: 17207 } },
    search3: { results: { matchCount: 23165 } },
    search4: { results: { matchCount: 24728 } },
    search5: { results: { matchCount: 41029 } },
    search6: { results: { matchCount: 41029 } },
}

export const MIGRATION_TO_GQL_INSIGHT_COMMITS_FIXTURE: Record<string, BulkSearchCommits> = {
    search0: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '2a1cd8a30c72780ad884159161d0ec828cfe69a3',
                        committer: {
                            date: '2020-05-29T19:48:57Z',
                        },
                    },
                },
            ],
        },
    },
    search1: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '68afed3a2812a197096720c80df928eba0ea0703',
                        committer: {
                            date: '2020-07-29T20:36:59Z',
                        },
                    },
                },
            ],
        },
    },
    search2: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: 'f62cb0864d367cfd09fb6f755807e6c25b44e6dd',
                        committer: {
                            date: '2020-09-29T20:22:52Z',
                        },
                    },
                },
            ],
        },
    },
    search3: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: 'ccede06037725365c3391e36b9d90c85eb00b71a',
                        committer: {
                            date: '2020-11-29T20:27:10Z',
                        },
                    },
                },
            ],
        },
    },
    search4: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: 'b29a72431e10adac2267cd4e5097f11d517e9139',
                        committer: {
                            date: '2021-01-31T00:52:56Z',
                        },
                    },
                },
            ],
        },
    },
    search5: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '4e565e36dc880f75c12982e2aba41f2445eeb4e1',
                        committer: {
                            date: '2021-03-30T20:47:08Z',
                        },
                    },
                },
            ],
        },
    },
    search6: {
        results: {
            results: [
                {
                    __typename: 'CommitSearchResult',
                    commit: {
                        oid: '1c2601a76662f6a0700b614a4a68406335075c29',
                        committer: {
                            date: '2021-05-31T17:05:12Z',
                        },
                    },
                },
            ],
        },
    },
}

export const MIGRATION_TO_GQL_INSIGHT_MATCHES_DATA_FIXTURE: Record<string, BulkSearchFields> = {
    search0: {
        results: {
            matchCount: 256,
        },
    },
    search1: {
        results: {
            matchCount: 254,
        },
    },
    search2: {
        results: {
            matchCount: 182,
        },
    },
    search3: {
        results: {
            matchCount: 179,
        },
    },
    search4: {
        results: {
            matchCount: 139,
        },
    },
    search5: {
        results: {
            matchCount: 139,
        },
    },
    search6: {
        results: {
            matchCount: 130,
        },
    },
    search7: {
        results: {
            matchCount: 0,
        },
    },
    search8: {
        results: {
            matchCount: 27,
        },
    },
    search9: {
        results: {
            matchCount: 208,
        },
    },
    search10: {
        results: {
            matchCount: 258,
        },
    },
    search11: {
        results: {
            matchCount: 340,
        },
    },
    search12: {
        results: {
            matchCount: 359,
        },
    },
    search13: {
        results: {
            matchCount: 422,
        },
    },
}

/**
 * For code-stats insight we have 1 step just in time processing. This fixture provides mock
 * data for this single step.
 */
export const SOURCEGRAPH_LANG_STATS_INSIGHT_DATA_FIXTURE: SearchResultsStatsResult = {
    search: {
        results: { limitHit: false },
        stats: {
            languages: [
                { name: 'Go', totalLines: 4498 },
                { name: 'JavaScript', totalLines: 948 },
                { name: 'Markdown', totalLines: 557 },
                { name: 'CSS', totalLines: 338 },
                { name: 'HTML', totalLines: 48 },
                { name: 'Text', totalLines: 21 },
                { name: 'YAML', totalLines: 1 },
            ],
        },
    },
}
