import { getInsightsDashboards } from './use-dashboards'

describe('getInsightsDashboards', () => {
    describe('should return dashboard list', () => {
        test('with personal (user-wide) dashboards only', () => {
            expect(
                getInsightsDashboards([
                    {
                        subject: {
                            __typename: 'User',
                            id: '101',
                            username: 'emirkusturica',
                            displayName: 'Emir Kusturica',
                            viewerCanAdminister: true,
                        },
                        settings: {
                            'insight.dashboard.testDashboard': {
                                title: 'Test Dashboard',
                                insightIds: ['insightID1', 'insightID2'],
                            },
                            'insight.dashboard.anotherTestDashboard': {
                                title: 'Another Test Dashboard',
                                insightIds: ['insightID3', 'insightID4'],
                            },
                        },
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                {
                    id: 'insight.dashboard.testDashboard',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    visibility: '101',
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
                {
                    id: 'insight.dashboard.anotherTestDashboard',
                    title: 'Another Test Dashboard',
                    insightIds: ['insightID3', 'insightID4'],
                    visibility: '101',
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
            ])
        })

        test('with org-wide and personal dashboards', () => {
            expect(
                getInsightsDashboards([
                    {
                        subject: {
                            __typename: 'Org',
                            id: '102',
                            name: 'sourcegraph',
                            displayName: 'Sourcegraph',
                            viewerCanAdminister: true,
                        },
                        settings: {
                            'insight.dashboard.testDashboard': {
                                title: 'Test Dashboard',
                                insightIds: ['insightID1', 'insightID2'],
                            },
                        },
                        lastID: null,
                    },
                    {
                        subject: {
                            __typename: 'User',
                            id: '101',
                            username: 'emirkusturica',
                            displayName: 'Emir Kusturica',
                            viewerCanAdminister: true,
                        },
                        settings: {
                            'insight.dashboard.anotherTestDashboard': {
                                title: 'Another Test Dashboard',
                                insightIds: ['insightID3', 'insightID4'],
                            },
                        },
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                {
                    id: 'insight.dashboard.testDashboard',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    visibility: '102',
                    owner: {
                        id: '102',
                        name: 'Sourcegraph',
                    },
                },
                {
                    id: 'insight.dashboard.anotherTestDashboard',
                    title: 'Another Test Dashboard',
                    insightIds: ['insightID3', 'insightID4'],
                    visibility: '101',
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
            ])
        })
    })
})
