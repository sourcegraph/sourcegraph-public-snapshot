import { ALL_INSIGHTS_DASHBOARD } from '../../../../core/types'

import { getInsightsDashboards } from './use-dashboards'

describe('getInsightsDashboards', () => {
    describe('should return empty custom list', () => {
        test('with null subject value', () => {
            expect(getInsightsDashboards(null)).toStrictEqual([])
        })

        test('with error like settings value', () => {
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
                        settings: new Error(),
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                // Even with empty or errored value of user settings we still have
                // generic built in insights dashboard - "All"
                ALL_INSIGHTS_DASHBOARD,
            ])
        })
    })

    describe('should return dashboard list', () => {
        test('with built in dashboard only if dashboard settings are empty', () => {
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
                        settings: {},
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
                        settings: {},
                        lastID: null,
                    },
                ])
            ).toStrictEqual([
                ALL_INSIGHTS_DASHBOARD,
                {
                    type: 'built-in',
                    insightIds: [],
                    owner: {
                        id: '102',
                        name: 'Sourcegraph',
                    },
                },
                {
                    type: 'built-in',
                    insightIds: [],
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
            ])
        })

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
                ALL_INSIGHTS_DASHBOARD,
                {
                    type: 'built-in',
                    insightIds: [],
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
                {
                    type: 'custom',
                    id: '001',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
                {
                    type: 'custom',
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
                ALL_INSIGHTS_DASHBOARD,
                {
                    type: 'built-in',
                    insightIds: [],
                    owner: {
                        id: '102',
                        name: 'Sourcegraph',
                    },
                },
                {
                    type: 'built-in',
                    insightIds: [],
                    owner: {
                        id: '101',
                        name: 'Emir Kusturica',
                    },
                },
                {
                    type: 'custom',
                    id: '001',
                    title: 'Test Dashboard',
                    insightIds: ['insightID1', 'insightID2'],
                    owner: {
                        id: '102',
                        name: 'Sourcegraph',
                    },
                },
                {
                    type: 'custom',
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
