import { Settings, SettingsCascade } from '@sourcegraph/shared/src/settings/settings'

import { diffCodeInsightsSettings, getGroupedStepSizes, getInsightsGroupedByType } from './analytics'

describe('Code Insight Analytics', () => {
    describe('getInsightsGroupedByType()', () => {
        test('with insights in org and in user settings', () => {
            const SETTINGS_CASCADE: SettingsCascade<Settings> = {
                subjects: [
                    {
                        subject: {
                            __typename: 'User',
                            displayName: 'User Name',
                            username: 'username',
                            id: 'user subject id',
                            viewerCanAdminister: true,
                        },
                        settings: {
                            'codeStatsInsights.insight.metabaseSearchVisibilityTest': {
                                title: 'Metabase search visibility test',
                                repository: 'github.com/metabase/metabase',
                                otherThreshold: 0.03,
                            },
                            'searchInsights.insight.selectRepoTest': {
                                title: 'select:repo test',
                                repositories: ['github.com/sourcegraph/sourcegraph'],
                                series: [
                                    {
                                        name: 'Test',
                                        query: 'String select:repo',
                                        stroke: 'var(--oc-red-7)',
                                    },
                                ],
                                step: {
                                    weeks: 6,
                                },
                            },
                        },
                        lastID: null,
                    },
                    {
                        subject: {
                            __typename: 'Org',
                            displayName: 'Org Name',
                            name: 'sourcegraph',
                            id: 'Org ID',
                            viewerCanAdminister: true,
                        },
                        settings: {
                            'codeStatsInsights.insight.ORGmetabaseSearchVisibilityTest': {
                                title: 'Metabase search visibility test',
                                repository: 'github.com/metabase/metabase',
                                otherThreshold: 0.03,
                            },
                            'searchInsights.insight.ORGselectRepoTest': {
                                title: 'select:repo test',
                                repositories: ['github.com/sourcegraph/sourcegraph'],
                                series: [
                                    {
                                        name: 'Test',
                                        query: 'String select:repo',
                                        stroke: 'var(--oc-red-7)',
                                    },
                                ],
                                step: {
                                    weeks: 6,
                                },
                            },
                        },
                        lastID: null,
                    },
                ],
                final: {},
            }

            expect(getInsightsGroupedByType(SETTINGS_CASCADE)).toStrictEqual({
                codeStatsInsights: 1,
                searchBasedInsights: 1,
                searchBasedBackendInsights: 0,
                searchBasedExtensionInsights: 1,
            })
        })
    })

    describe('getGroupedStepSizes()', () => {
        test('with array of search and code stats insight', () => {
            const settings = {
                'searchInsights.insight.name1': {
                    step: { years: 2 },
                },
                'searchInsights.insight.name2': {
                    step: { months: 5 },
                },
                'searchInsights.insight.name3': {
                    step: { weeks: 6 },
                },
                'searchInsights.insight.name4': {
                    step: { days: 6 },
                },
                'codeStatsInsights.insight.ORGmetabaseSearchVisibilityTest': {
                    title: 'Metabase search visibility test',
                    repository: 'github.com/metabase/metabase',
                    otherThreshold: 0.03,
                },
            }
            expect(getGroupedStepSizes(settings)).toStrictEqual([365 * 2, 30 * 5, 7 * 6, 6])
        })
    })
    describe('diffCodeInsightsSettings()', () => {
        test('addition', () => {
            const oldSettings = {}
            const newSettings = {
                'codeStatsInsights.insight.Title': { query: 'repo:^github\\.com/sourcegraph/sourcegraph$' },
            }

            expect(diffCodeInsightsSettings(oldSettings, newSettings)).toStrictEqual([
                { action: 'Addition', insightType: 'codeStatsInsights' },
            ])
        })

        test('removal', () => {
            const oldSettings = {
                'codeStatsInsights.insight.Title': { query: 'repo:^github\\.com/sourcegraph/sourcegraph$' },
            }
            const newSettings = {}

            expect(diffCodeInsightsSettings(oldSettings, newSettings)).toStrictEqual([
                { action: 'Removal', insightType: 'codeStatsInsights' },
            ])
        })

        test('no changes', () => {
            const oldSettings = { 'codeStatsInsights.query': 'repo:^github\\.com/sourcegraph/sourcegraph$' }
            const newSettings = { 'codeStatsInsights.query': 'repo:^github\\.com/sourcegraph/sourcegraph$' }

            expect(diffCodeInsightsSettings(oldSettings, newSettings)).toStrictEqual([])
        })

        test('edit', () => {
            const oldSettings = {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'Migration to new GraphQL TS types',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [
                        {
                            name: 'Imports of old GQL.* types',
                            query: 'patternType:regex case:yes \\*\\sas\\sGQL',
                            stroke: 'var(--oc-red-7)',
                        },
                        {
                            name: 'Imports of new graphql-operations types',
                            query: "patternType:regexp case:yes /graphql-operations'",
                            stroke: 'var(--oc-blue-7)',
                        },
                    ],
                    step: {
                        weeks: 6,
                    },
                },
            }
            // user changed step to 5 weeks
            const newSettings = {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'Migration to new GraphQL TS types',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [
                        {
                            name: 'Imports of old GQL.* types',
                            query: 'patternType:regex case:yes \\*\\sas\\sGQL',
                            stroke: 'var(--oc-red-7)',
                        },
                        {
                            name: 'Imports of new graphql-operations types',
                            query: "patternType:regexp case:yes /graphql-operations'",
                            stroke: 'var(--oc-blue-7)',
                        },
                    ],
                    step: {
                        weeks: 5,
                    },
                },
            }

            expect(diffCodeInsightsSettings(oldSettings, newSettings)).toStrictEqual([
                { action: 'Edit', insightType: 'searchInsights' },
            ])
        })
    })
})
