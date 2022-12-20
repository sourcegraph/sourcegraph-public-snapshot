import { waitFor } from '@testing-library/react'
import * as H from 'history'
import { of } from 'rxjs'

import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { SiteUsagePeriodFields } from '../../graphql-operations'

import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'

describe('SiteAdminOverviewPage', () => {
    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        overviewComponents: [],
    }

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
        const usageStat: SiteUsagePeriodFields = {
            __typename: 'SiteUsagePeriod',
            userCount: 10,
            registeredUserCount: 8,
            anonymousUserCount: 2,
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
                        waus: [usageStat, usageStat],
                    })
                }
            />
        )
        // ensure the hooks ran and the "API response" has been received
        await waitFor(() => expect(component.asFragment()).toMatchSnapshot())
    })
})
