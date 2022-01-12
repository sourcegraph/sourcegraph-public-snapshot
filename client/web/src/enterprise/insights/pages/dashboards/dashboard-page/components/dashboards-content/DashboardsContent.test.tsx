import { useApolloClient } from '@apollo/client'
import { MockedResponse } from '@apollo/client/testing'
import { waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import sinon from 'sinon'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import { AuthenticatedUser } from '../../../../../../../auth'
import { InsightsDashboardsResult } from '../../../../../../../graphql-operations'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../../../../../core/backend/gql-api/code-insights-gql-backend'
import { GET_DASHBOARD_INSIGHTS_GQL } from '../../../../../core/backend/gql-api/gql/GetDashboardInsights'
import { GET_INSIGHTS_DASHBOARDS_GQL } from '../../../../../core/backend/gql-api/gql/GetInsightsDashboards'
import { GET_INSIGHTS_SUBJECTS_GQL } from '../../../../../core/backend/gql-api/gql/GetInsightSubjects'

import { DashboardsContent } from './DashboardsContent'

// This mocked user is used internally to display DashboardSelect
const mockUser: Partial<AuthenticatedUser> = { id: 'user-foo', username: 'userfoo', organizations: { nodes: [] } }

jest.mock('@sourcegraph/web/src/auth', () => ({
    authenticatedUser: {
        subscribe: ({ next }: { next: (args: unknown) => unknown }) => {
            next(mockUser)
            return { unsubscribe: () => null }
        },
    },
}))

const mockTelemetryService = {
    log: sinon.spy(),
    logViewEvent: sinon.spy(),
}

const Wrapper: React.FunctionComponent = ({ children }) => {
    const apolloClient = useApolloClient()
    const api = new CodeInsightsGqlBackend(apolloClient)

    return <CodeInsightsBackendContext.Provider value={api}>{children}</CodeInsightsBackendContext.Provider>
}

const mockDashboard: InsightsDashboardsResult['insightsDashboards']['nodes'][0] = {
    id: 'foo',
    title: 'Global Dashboard',
    views: null,
    grants: {
        users: [],
        organizations: [],
        global: true,
    },
}

const mockDashboard2: InsightsDashboardsResult['insightsDashboards']['nodes'][0] = {
    id: 'bar',
    title: 'Global Dashboard 2',
    views: null,
    grants: {
        users: [],
        organizations: [],
        global: true,
    },
}

const mocks: MockedResponse[] = [
    {
        request: {
            query: GET_INSIGHTS_DASHBOARDS_GQL,
            variables: { id: undefined },
        },
        result: {
            data: { insightsDashboards: { nodes: [mockDashboard, mockDashboard2] } },
        },
    },
    {
        request: {
            query: GET_INSIGHTS_DASHBOARDS_GQL,
            variables: { id: 'foo' },
        },
        result: {
            data: { insightsDashboards: { nodes: [mockDashboard, mockDashboard2] } },
        },
    },
    {
        request: {
            query: GET_INSIGHTS_SUBJECTS_GQL,
        },
        result: {
            data: {
                currentUser: null,
                site: null,
            },
        },
    },
    {
        request: {
            query: GET_DASHBOARD_INSIGHTS_GQL,
            variables: { id: 'foo' },
        },
        result: {
            data: {
                insightsDashboards: {
                    nodes: [
                        {
                            id: 'foo',
                            views: null,
                        },
                    ],
                },
            },
        },
    },
]

const renderDashboardsContent = (component: React.ReactElement): RenderWithRouterResult =>
    renderWithRouter(
        <MockedTestProvider mocks={mocks}>
            <Wrapper>{component}</Wrapper>
        </MockedTestProvider>
    )

beforeEach(() => {
    jest.clearAllMocks()
})

describe('DashboardsContent', () => {
    it('renders a loading indicator', () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="baz" telemetryService={mockTelemetryService} />
        )

        expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    })

    it('renders dashboard not found', async () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="baz" telemetryService={mockTelemetryService} />
        )

        await waitFor(() => expect(screen.getByText("Hmm, the dashboard wasn't found.")).toBeInTheDocument())
    })

    it('renders a dashboard', async () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="foo" telemetryService={mockTelemetryService} />
        )

        await waitFor(() => expect(screen.getByRole('button', { name: /Global Dashboard/ })).toBeInTheDocument())
    })

    it('redirect to new dashboard page on selection', async () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="foo" telemetryService={mockTelemetryService} />
        )
        const { history } = screen

        const chooseDashboard = await waitFor(() => screen.getByRole('button', { name: /Choose a dashboard/ }))
        userEvent.click(chooseDashboard)

        const dashboard2 = screen.getByRole('option', { name: /Global Dashboard 2/ })
        userEvent.click(dashboard2)

        expect(history.location.pathname).toEqual('/insights/dashboards/bar')
    })

    // Note: the rest of these are unwritten due to a bug in ReachUI.
    // You cannot trigger the `onSelect` programmatically.
    // https://github.com/reach/reach-ui/issues/886

    // it('redirects to dashboard edit page', () => {
    //     const { history } = renderWithRouter(
    //         <DashboardsContent dashboardID="foo" telemetryService={mockTelemetryService} />
    //     )

    //     const dashboardMenu = screen.getByRole('button', { name: /Dashboard options/ })
    //     userEvent.click(dashboardMenu)

    //     const editDashboard = screen.getByRole('menuitem', { name: /Configure dashboard/ })
    //     userEvent.click(editDashboard)

    //     expect(history.location.pathname).toEqual('/insights/dashboards/foo/edit')
    // })

    // opens add insight modal

    // opens delete dashboard modal

    // copies dashboard url
})
