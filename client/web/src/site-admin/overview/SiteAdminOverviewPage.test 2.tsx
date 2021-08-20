import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'
import sinon from 'sinon'

import { ISiteUsagePeriod } from '@sourcegraph/shared/src/graphql/schema'

import { PageTitle } from '../../components/PageTitle'

import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'

describe('SiteAdminOverviewPage', () => {
    afterEach(() => {
        PageTitle.titleSet = false
    })

    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        overviewComponents: [],
    }

    test('activation in progress', done => {
        const component = renderer.create(
            <SiteAdminOverviewPage
                {...baseProps}
                activation={{
                    steps: [
                        {
                            id: 'ConnectedCodeHost' as const,
                            title: 'Add repositories',
                            detail: 'Configure Sourcegraph to talk to your code host',
                        },
                    ],
                    completed: {
                        ConnectedCodeHost: false,
                    },
                    update: sinon.stub(),
                    refetch: sinon.stub(),
                }}
                _fetchOverview={() =>
                    of({
                        repositories: 0,
                        repositoryStats: {
                            gitDirBytes: '1825299556',
                            indexedLinesCount: '2616264',
                        },
                        users: 1,
                        orgs: 1,
                        surveyResponses: {
                            totalCount: 0,
                            averageScore: 0,
                        },
                    })
                }
                _fetchWeeklyActiveUsers={() =>
                    of({
                        __typename: 'SiteUsageStatistics',
                        daus: [],
                        waus: [],
                        maus: [],
                    })
                }
            />
        )
        // ensure the hooks ran and the "API response" has been received
        setTimeout(() => {
            expect(component.toJSON()).toMatchSnapshot()
            done()
        })
    })

    test('< 2 users', done => {
        const component = renderer.create(
            <SiteAdminOverviewPage
                {...baseProps}
                _fetchOverview={() =>
                    of({
                        repositories: 100,
                        repositoryStats: {
                            gitDirBytes: '1825299556',
                            indexedLinesCount: '2616264',
                        },
                        users: 1,
                        orgs: 1,
                        surveyResponses: {
                            totalCount: 1,
                            averageScore: 10,
                        },
                    })
                }
                _fetchWeeklyActiveUsers={() =>
                    of({
                        __typename: 'SiteUsageStatistics',
                        daus: [],
                        waus: [],
                        maus: [],
                    })
                }
            />
        )
        // ensure the hooks ran and the "API response" has been received
        setTimeout(() => {
            expect(component.toJSON()).toMatchSnapshot()
            done()
        })
    })
    test('>= 2 users', done => {
        const usageStat: ISiteUsagePeriod = {
            __typename: 'SiteUsagePeriod',
            userCount: 10,
            registeredUserCount: 8,
            anonymousUserCount: 2,
            integrationUserCount: 0,
            startTime: new Date().toISOString(),
        }
        const component = renderer.create(
            <SiteAdminOverviewPage
                {...baseProps}
                _fetchOverview={() =>
                    of({
                        repositories: 100,
                        repositoryStats: {
                            gitDirBytes: '1825299556',
                            indexedLinesCount: '2616264',
                        },
                        users: 10,
                        orgs: 5,
                        surveyResponses: {
                            totalCount: 100,
                            averageScore: 10,
                        },
                    })
                }
                _fetchWeeklyActiveUsers={() =>
                    of({
                        __typename: 'SiteUsageStatistics',
                        daus: [],
                        waus: [usageStat, usageStat],
                        maus: [],
                    })
                }
            />
        )
        // ensure the hooks ran and the "API response" has been received
        setTimeout(() => {
            expect(component.toJSON()).toMatchSnapshot()
            done()
        })
    })
})
