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
import { GET_INSIGHTS_GQL } from '../../../../../core/backend/gql-api/gql/GetInsights'
import { GET_INSIGHTS_DASHBOARDS_GQL } from '../../../../../core/backend/gql-api/gql/GetInsightsDashboards'
import { GET_INSIGHTS_SUBJECTS_GQL } from '../../../../../core/backend/gql-api/gql/GetInsightSubjects'

import { DashboardsContent } from './DashboardsContent'

type UserEvent = typeof userEvent

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

const mockCopyURL = sinon.spy()

jest.mock('./hooks/use-copy-url-handler', () => ({
    useCopyURLHandler: () => [mockCopyURL],
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
    {
        request: {
            query: GET_INSIGHTS_GQL,
            variables: {},
        },
        result: {
            data: { insightViews: { nodes: [] } },
        },
    },
]

const renderDashboardsContent = (dashboardID: string = 'foo'): RenderWithRouterResult & { user: UserEvent } => ({
    user: userEvent,
    ...renderWithRouter(
        <MockedTestProvider mocks={mocks}>
            <Wrapper>
                <DashboardsContent dashboardID={dashboardID} telemetryService={mockTelemetryService} />
            </Wrapper>
        </MockedTestProvider>
    ),
})

const triggerDashboardMenuItem = async (screen: RenderWithRouterResult & { user: UserEvent }, name: RegExp) => {
    const { user } = screen
    const dashboardMenu = await waitFor(() => screen.getByRole('button', { name: /Dashboard options/ }))
    user.click(dashboardMenu)

    const dashboardMenuItem = screen.getByRole('menuitem', { name })

    // We're simulating keyboard navigation here to circumvent a bug in ReachUI
    // ReachUI does not respond to programmatic click events on menu items
    dashboardMenuItem.focus()
    user.keyboard(' ')
}

beforeEach(() => {
    jest.clearAllMocks()
})

describe('DashboardsContent', () => {
    it('renders a loading indicator', () => {
        const screen = renderDashboardsContent('baz')

        expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    })

    it('renders dashboard not found', async () => {
        const screen = renderDashboardsContent('baz')

        await waitFor(() => expect(screen.getByText("Hmm, the dashboard wasn't found.")).toBeInTheDocument())
    })

    it('renders a dashboard', async () => {
        const screen = renderDashboardsContent()

        await waitFor(() => expect(screen.getByRole('button', { name: /Global Dashboard/ })).toBeInTheDocument())
    })

    it('redirect to new dashboard page on selection', async () => {
        const screen = renderDashboardsContent()
        const { history, user } = screen

        const chooseDashboard = await waitFor(() => screen.getByRole('button', { name: /Choose a dashboard/ }))
        user.click(chooseDashboard)

        const dashboard2 = screen.getByRole('option', { name: /Global Dashboard 2/ })
        user.click(dashboard2)

        expect(history.location.pathname).toEqual('/insights/dashboards/bar')
    })

    it('redirects to dashboard edit page', async () => {
        const screen = renderDashboardsContent()

        const { history } = screen

        await triggerDashboardMenuItem(screen, /Configure dashboard/)

        expect(history.location.pathname).toEqual('/insights/dashboards/foo/edit')
    })

    it('opens add insight modal', async () => {
        const screen = renderDashboardsContent()

        await triggerDashboardMenuItem(screen, /Add or remove insights/)

        const addInsightHeader = await waitFor(() =>
            screen.getByRole('heading', { name: /Add insight to Global Dashboard/ })
        )
        expect(addInsightHeader).toBeInTheDocument()
    })

    it('opens delete dashboard modal', async () => {
        const screen = renderDashboardsContent()

        await triggerDashboardMenuItem(screen, /Delete/)

        const addInsightHeader = await waitFor(() => screen.getByRole('heading', { name: /Delete/ }))
        expect(addInsightHeader).toBeInTheDocument()
    })

    // copies dashboard url
    it('copies dashboard url', async () => {
        const screen = renderDashboardsContent()

        await triggerDashboardMenuItem(screen, /Copy link/)

        sinon.assert.calledOnce(mockCopyURL)
    })
})
