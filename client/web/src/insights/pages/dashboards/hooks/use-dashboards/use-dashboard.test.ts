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
                            'insights.dashboard.testDashboard': {
                                title: 'Test Dashboard',
                                insightIds: ['insightID1', 'insightID2'],
                            },
                            'insights.dashboard.anotherTestDashboard': {
                                title: 'Another Test Dashboard',
                                insightIds: ['insightID3', 'insightID4'],
                            },
                        },
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                {
                    id: 'insights.dashboard.testDashboard',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    visibility: '101',
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
                {
                    id: 'insights.dashboard.anotherTestDashboard',
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
                            'insights.dashboard.testDashboard': {
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
                            'insights.dashboard.anotherTestDashboard': {
                                title: 'Another Test Dashboard',
                                insightIds: ['insightID3', 'insightID4'],
                            },
                        },
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                {
                    id: 'insights.dashboard.testDashboard',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    visibility: '102',
                    owner: {
                        id: '102',
                        name: 'Sourcegraph',
                    },
                },
                {
                    id: 'insights.dashboard.anotherTestDashboard',
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
