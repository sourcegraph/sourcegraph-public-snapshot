import { waitFor } from '@testing-library/react'
import * as H from 'history'
import { of } from 'rxjs'
import sinon from 'sinon'

import { ISiteUsagePeriod } from '@sourcegraph/shared/src/schema'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'

describe('SiteAdminOverviewPage', () => {
    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        overviewComponents: [],
    }

    test('activation in progress', async () => {
        const component = renderWithBrandedContext(
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
        await waitFor(() => expect(component.asFragment()).toMatchSnapshot())
    })

    test('< 2 users', async () => {
        const component = renderWithBrandedContext(
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
        await waitFor(() => expect(component.asFragment()).toMatchSnapshot())
    })

    test('>= 2 users', async () => {
        const usageStat: ISiteUsagePeriod = {
            __typename: 'SiteUsagePeriod',
            userCount: 10,
            registeredUserCount: 8,
            anonymousUserCount: 2,
            integrationUserCount: 0,
            startTime: new Date().toISOString(),
        }

        // Do this mock for resolving issue
        // `Error: Uncaught [TypeError: tspan.node(...).getComputedTextLength is not a function`
        // from `client/web/src/components/d3/BarChart.tsx`
        // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-member-access
        ;(window.SVGElement as any).prototype.getComputedTextLength = () => 500

        const component = renderWithBrandedContext(
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
        await waitFor(() => expect(component.asFragment()).toMatchSnapshot())
    })
})
