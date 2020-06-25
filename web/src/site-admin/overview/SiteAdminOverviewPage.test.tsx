import * as H from 'history'
import React from 'react'
import { of } from 'rxjs'
import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'
import { eventLogger } from '../../tracking/eventLogger'
import sinon from 'sinon'
import { ISiteUsagePeriod } from '../../../../shared/src/graphql/schema'
import { PageTitle } from '../../components/PageTitle'
import { mount } from 'enzyme'

jest.mock('../../components/d3/BarChart', () => ({ BarChart: 'BarChart' }))

describe('SiteAdminOverviewPage', () => {
    afterEach(() => {
        PageTitle.titleSet = false
    })

    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        overviewComponents: [],
    }

    let stub: sinon.SinonStub<[string, (boolean | undefined)?], void>

    beforeAll(() => {
        stub = sinon.stub(eventLogger, 'logViewEvent')
    })

    afterAll(() => {
        if (stub) {
            stub.restore()
        }
    })

    test('activation in progress', () => {
        expect(
            mount(
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
        ).toMatchSnapshot()
    })

    test('< 2 users', () => {
        expect(
            mount(
                <SiteAdminOverviewPage
                    {...baseProps}
                    _fetchOverview={() =>
                        of({
                            repositories: 100,
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
        ).toMatchSnapshot()
    })
    test('>= 2 users', () => {
        const usageStat: ISiteUsagePeriod = {
            __typename: 'SiteUsagePeriod',
            userCount: 10,
            registeredUserCount: 8,
            anonymousUserCount: 2,
            integrationUserCount: 0,
            startTime: new Date().toISOString(),
            stages: undefined as any,
        }
        const component = mount(
            <SiteAdminOverviewPage
                {...baseProps}
                _fetchOverview={() =>
                    of({
                        repositories: 100,
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
        component.update()
        expect(component).toMatchSnapshot()
    })
})
