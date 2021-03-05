import { diffCodeInsightsSettings } from './analytics'

describe('Code Insight Analytics', () => {
    describe('diffCodeInsightsSettings()', () => {
        test('addition', () => {
            const oldSettings = {}
            const newSettings = { 'codeStatsInsights.query': 'repo:^github\\.com/sourcegraph/sourcegraph$' }

            expect(diffCodeInsightsSettings(oldSettings, newSettings)).toStrictEqual([
                { action: 'Addition', insightType: 'codeStatsInsights' },
            ])
        })

        test('removal', () => {
            const oldSettings = { 'codeStatsInsights.query': 'repo:^github\\.com/sourcegraph/sourcegraph$' }
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
