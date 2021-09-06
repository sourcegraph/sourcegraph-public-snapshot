import { View } from 'sourcegraph'

import { BulkSearchCommits } from '../../../graphql-operations';

export const SEARCH_INSIGHT_COMMITS_MOCK: Record<string, BulkSearchCommits> = {
    'search0': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': '0b81b624c24f5fc4d53fd10651eb84d67072e74e',
                        'committer': {
                            'date': '2009-09-04T05:00:34Z'
                        }
                    }
                }
            ]
        }
    },
    'search1': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': '03956c5dde9dcb22a53e7d3f259a0e98dd50704b',
                        'committer': {
                            'date': '2011-09-05T19:10:27Z'
                        }
                    }
                }
            ]
        }
    },
    'search2': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': '948698c42a2a05b1ebebf3cda5945b0764426f57',
                        'committer': {
                            'date': '2013-09-04T22:25:47Z'
                        }
                    }
                }
            ]
        }
    },
    'search3': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': 'db37595dc3a4ceb7953dd89b2122a974ff70b311',
                        'committer': {
                            'date': '2015-09-05T18:20:08Z'
                        }
                    }
                }
            ]
        }
    },
    'search4': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': '4472874fb5fddb44d402f4fdf064149b68dce68c',
                        'committer': {
                            'date': '2017-09-05T16:16:25Z'
                        }
                    }
                }
            ]
        }
    },
    'search5': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': '6f34ec9377091a1665d554a54bc58812813c7c1a',
                        'committer': {
                            'date': '2019-07-09T19:59:43Z'
                        }
                    }
                }
            ]
        }
    },
    'search6': {
        'results': {
            'results': [
                {
                    'commit': {
                        'oid': '6f34ec9377091a1665d554a54bc58812813c7c1a',
                        'committer': {
                            'date': '2019-07-09T19:59:43Z'
                        }
                    }
                }
            ]
        }
    }
}

export const SEARCH_INSIGHT_RESULT_MOCK = {
    'search0': { 'results': { 'matchCount': 10245 } },
    'search1': { 'results': { 'matchCount': 16502 } },
    'search2': { 'results': { 'matchCount': 17207 } },
    'search3': { 'results': { 'matchCount': 23165 } },
    'search4': { 'results': { 'matchCount': 24728 } },
    'search5': { 'results': { 'matchCount': 41029 } },
    'search6': { 'results': { 'matchCount': 41029 } }
}

export const CODE_STATS_RESULT_MOCK = {
    'search': {
        'results': { 'limitHit': false },
        'stats': {
            'languages': [
                { 'name': 'Go', 'totalLines': 4498 },
                { 'name': 'JavaScript', 'totalLines': 948 },
                { 'name': 'Markdown', 'totalLines': 557 },
                { 'name': 'CSS', 'totalLines': 338 },
                { 'name': 'HTML', 'totalLines': 48 },
                { 'name': 'Text', 'totalLines': 21 },
                { 'name': 'YAML', 'totalLines': 1 }
            ]
        }
    }
}

/**
 * Mock result data of search based insight extension - team size,
 * See https://github.com/sourcegraph/sourcegraph-search-insights
 * basically line chart with one grown data series.
 * */
export const INSIGHT_VIEW_TEAM_SIZE: View = {
    title: 'Team size',
    content: [
        {
            chart: 'line',
            data: [
                { date: 1574802000000, 'team members': 22 },
                { date: 1582750800000, 'team members': 29 },
                { date: 1590526800000, 'team members': 40 },
                { date: 1598475600000, 'team members': 58 },
                { date: 1606424400000, 'team members': 65 },
                { date: 1614373200000, 'team members': 97 },
                { date: 1622062800000, 'team members': 100 },
            ],
            series: [
                {
                    dataKey: 'team members',
                    name: 'team members',
                    stroke: 'var(--oc-teal-7)',
                },
            ],
            xAxis: { dataKey: 'date', type: 'number', scale: 'time' },
        },
    ],
}

/**
 * Mock result of search based insight extension types migration
 * See https://github.com/sourcegraph/sourcegraph-search-insights
 * Line chart example with two data series.
 * */
export const INSIGHT_VIEW_TYPES_MIGRATION: View = {
    title: 'Migration to new GraphQL TS types',
    content: [
        {
            chart: 'line',
            data: [
                {
                    date: 1600203600000,
                    'Imports of old GQL.* types': 188,
                    'Imports of new graphql-operations types': 203,
                },
                {
                    date: 1603832400000,
                    'Imports of old GQL.* types': 178,
                    'Imports of new graphql-operations types': 234,
                },
                {
                    date: 1607461200000,
                    'Imports of old GQL.* types': 162,
                    'Imports of new graphql-operations types': 282,
                },
                {
                    date: 1611090000000,
                    'Imports of old GQL.* types': 139,
                    'Imports of new graphql-operations types': 340,
                },
                {
                    date: 1614718800000,
                    'Imports of old GQL.* types': 139,
                    'Imports of new graphql-operations types': 354,
                },
                {
                    date: 1618347600000,
                    'Imports of old GQL.* types': 139,
                    'Imports of new graphql-operations types': 369,
                },
                {
                    date: 1621976400000,
                    'Imports of old GQL.* types': 131,
                    'Imports of new graphql-operations types': 427,
                },
            ],
            series: [
                {
                    dataKey: 'Imports of old GQL.* types',
                    name: 'Imports of old GQL.* types',
                    stroke: 'var(--oc-red-7)',
                },
                {
                    dataKey: 'Imports of new graphql-operations types',
                    name: 'Imports of new graphql-operations types',
                    stroke: 'var(--oc-blue-7)',
                },
            ],
            xAxis: { dataKey: 'date', type: 'number', scale: 'time' },
        },
    ],
}

/**
 * Mock result from code stats insight extension
 * See https://github.com/sourcegraph/sourcegraph-code-stats-insights
 * Pie chart with language usage.
 * */
export const CODE_STATS_INSIGHT_LANG_USAGE: View = {
    title: 'Example sourcegraph lang usage',
    content: [
        {
            chart: 'pie',
            pies: [
                {
                    data: [
                        {
                            name: 'Go',
                            totalLines: 392208,
                            fill: '#00ADD8',
                        },
                        {
                            name: 'HTML',
                            totalLines: 225121,
                            fill: '#e34c26',
                        },
                        {
                            name: 'TypeScript',
                            totalLines: 201214,
                            fill: '#2b7489',
                        },
                        {
                            name: 'YAML',
                            totalLines: 163177,
                            fill: '#cb171e',
                        },
                        {
                            name: 'JSON',
                            totalLines: 107794,
                            fill: 'gray',
                        },
                        {
                            name: 'Markdown',
                            totalLines: 52111,
                            fill: '#083fa1',
                        },
                        {
                            name: 'Other',
                            totalLines: 60779,
                            fill: 'gray',
                        },
                    ],
                    dataKey: 'totalLines',
                    nameKey: 'name',
                    fillKey: 'fill',
                    linkURLKey: 'linkURL',
                },
            ],
        },
    ],
}

/**
 * Backend Insight with linear increasing data
 */
export const LINEAR_BACKEND_INSIGHT = {
    id: 'backend_ID_001',
    title: 'Testing Insight',
    description: 'Insight for testing',
    series: [
        {
            label: 'Insight',
            points: [
                {
                    dateTime: '2021-02-11T00:00:00Z',
                    value: 9,
                },
                {
                    dateTime: '2021-01-27T00:00:00Z',
                    value: 8,
                },
                {
                    dateTime: '2021-01-12T00:00:00Z',
                    value: 7,
                },
                {
                    dateTime: '2020-12-28T00:00:00Z',
                    value: 6,
                },
                {
                    dateTime: '2020-12-13T00:00:00Z',
                    value: 5,
                },
                {
                    dateTime: '2020-11-28T00:00:00Z',
                    value: 4,
                },
                {
                    dateTime: '2020-11-13T00:00:00Z',
                    value: 3,
                },
                {
                    dateTime: '2020-10-29T00:00:00Z',
                    value: 2,
                },
                {
                    dateTime: '2020-10-14T00:00:00Z',
                    value: 1,
                },
                {
                    dateTime: '2020-09-29T00:00:00Z',
                    value: 0,
                },
            ],
            status: {
                pendingJobs: 0,
                completedJobs: 0,
                failedJobs: 0,
                backfillQueuedAt: '2021-02-11T00:00:00Z',
            },
        },
    ],
}

/**
 * Mock data for gql api our backend insights.
 */
export const BACKEND_INSIGHTS = [LINEAR_BACKEND_INSIGHT]

/**
 * Mock gql api for live preview chart of `INSIGHT_VIEW_TYPES_MIGRATION` insight.
 * {@link INSIGHT_VIEW_TYPES_MIGRATION}
 */
export const INSIGHT_TYPES_MIGRATION_COMMITS = {
    search0: {
        results: {
            results: [
                {
                    commit: {
                        oid: '2a1cd8a30c72780ad884159161d0ec828cfe69a3',
                        committer: {
                            date: '2020-05-30T19:48:57Z',
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
                    commit: {
                        oid: '68afed3a2812a197096720c80df928eba0ea0703',
                        committer: {
                            date: '2020-07-31T20:36:59Z',
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
                    commit: {
                        oid: 'f62cb0864d367cfd09fb6f755807e6c25b44e6dd',
                        committer: {
                            date: '2020-09-30T20:22:52Z',
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
                    commit: {
                        oid: 'ccede06037725365c3391e36b9d90c85eb00b71a',
                        committer: {
                            date: '2020-11-30T20:27:10Z',
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
                    commit: {
                        oid: 'b29a72431e10adac2267cd4e5097f11d517e9139',
                        committer: {
                            date: '2021-01-30T00:52:56Z',
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
                    commit: {
                        oid: '4e565e36dc880f75c12982e2aba41f2445eeb4e1',
                        committer: {
                            date: '2021-03-31T20:47:08Z',
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

/**
 * Mock Bulk Search gql api for live preview chart of `INSIGHT_VIEW_TYPES_MIGRATION` insight.
 * */
export const INSIGHT_TYPES_MIGRATION_BULK_SEARCH = {
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
 * Code stats insight (gql query - LangStatsInsightContent) live preview mock.
 */
export const LangStatsInsightContent = {
    search: {
        results: {
            limitHit: false,
        },
        stats: {
            languages: [
                {
                    name: 'Markdown',
                    totalLines: 83176,
                },
                {
                    name: 'SVG',
                    totalLines: 17369,
                },
                {
                    name: 'YAML',
                    totalLines: 16226,
                },
                {
                    name: 'TypeScript',
                    totalLines: 9164,
                },
                {
                    name: 'SCSS',
                    totalLines: 2597,
                },
                {
                    name: 'JSON',
                    totalLines: 1801,
                },
                {
                    name: 'HTML',
                    totalLines: 1281,
                },
                {
                    name: 'CSS',
                    totalLines: 1188,
                },
                {
                    name: 'JavaScript',
                    totalLines: 473,
                },
                {
                    name: 'Go',
                    totalLines: 260,
                },
                {
                    name: 'Text',
                    totalLines: 174,
                },
                {
                    name: 'TOML',
                    totalLines: 106,
                },
                {
                    name: 'EditorConfig',
                    totalLines: 20,
                },
                {
                    name: 'Jsonnet',
                    totalLines: 15,
                },
                {
                    name: 'Makefile',
                    totalLines: 12,
                },
                {
                    name: 'Ignore List',
                    totalLines: 8,
                },
            ],
        },
    },
}
