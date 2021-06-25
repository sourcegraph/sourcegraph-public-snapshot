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
                            'insights.dashboards': {
                                'insights.dashboard.testDashboard': {
                                    id: '001',
                                    title: 'Test Dashboard',
                                    insightIds: ['insightID1', 'insightID2'],
                                },
                                'insights.dashboard.anotherTestDashboard': {
                                    id: '002',
                                    title: 'Another Test Dashboard',
                                    insightIds: ['insightID3', 'insightID4'],
                                },
                            },
                        },
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                {
                    id: '001',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
                {
                    id: '002',
                    title: 'Another Test Dashboard',
                    insightIds: ['insightID3', 'insightID4'],
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
                            'insights.dashboards': {
                                'insights.dashboard.testDashboard': {
                                    id: '001',
                                    title: 'Test Dashboard',
                                    insightIds: ['insightID1', 'insightID2'],
                                },
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
                            'insights.dashboards': {
                                'insights.dashboard.anotherTestDashboard': {
                                    id: '002',
                                    title: 'Another Test Dashboard',
                                    insightIds: ['insightID3', 'insightID4'],
                                },
                            },
                        },
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                {
                    id: '001',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    owner: {
                        id: '102',
                        name: 'Sourcegraph',
                    },
                },
                {
                    id: '002',
                    title: 'Another Test Dashboard',
                    insightIds: ['insightID3', 'insightID4'],
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
            ])
        })
    })
})
