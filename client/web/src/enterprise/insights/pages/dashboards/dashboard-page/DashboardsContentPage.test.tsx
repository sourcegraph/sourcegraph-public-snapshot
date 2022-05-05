import React from 'react'

import { useApolloClient } from '@apollo/client'
import { MockedResponse } from '@apollo/client/testing'
import { waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { renderWithBrandedContext, RenderWithBrandedContextResult } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/testing/MockIntersectionObserver'

import {
    GetAccessibleInsightsListResult,
    GetDashboardInsightsResult,
    GetInsightsResult,
    InsightsDashboardsResult,
    InsightSubjectsResult,
} from '../../../../../graphql-operations'
import { CodeInsightsBackendContext, CodeInsightsGqlBackend } from '../../../core'
import {
    GET_ACCESSIBLE_INSIGHTS_LIST,
    GET_DASHBOARD_INSIGHTS_GQL,
    GET_INSIGHTS_GQL,
    GET_INSIGHTS_DASHBOARDS_GQL,
    GET_INSIGHTS_DASHBOARD_OWNERS_GQL,
} from '../../../core/backend/gql-backend'

import { DashboardsContentPage } from './DashboardsContentPage'

type UserEvent = typeof userEvent

const mockCopyURL = sinon.spy()

jest.mock('../../../hooks/use-copy-url-handler', () => ({
    useCopyURLHandler: () => [mockCopyURL],
}))

const mockTelemetryService = {
    log: sinon.spy(),
    logViewEvent: sinon.spy(),
    logPageView: sinon.spy(),
}

const Wrapper: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => {
    const apolloClient = useApolloClient()
    const api = new CodeInsightsGqlBackend(apolloClient)

    return <CodeInsightsBackendContext.Provider value={api}>{children}</CodeInsightsBackendContext.Provider>
}

const mockDashboard: InsightsDashboardsResult['insightsDashboards']['nodes'][0] = {
    __typename: 'InsightsDashboard',
    id: 'foo',
    title: 'Global Dashboard',
    views: null,
    grants: {
        __typename: 'InsightsPermissionGrants',
        users: [],
        organizations: [],
        global: true,
    },
}

const mockDashboard2: InsightsDashboardsResult['insightsDashboards']['nodes'][0] = {
    __typename: 'InsightsDashboard',
    id: 'bar',
    title: 'Global Dashboard 2',
    views: null,
    grants: {
        __typename: 'InsightsPermissionGrants',
        users: [],
        organizations: [],
        global: true,
    },
}

const userMock = {
    __typename: 'User',
    id: '001',
    organizations: {
        nodes: [],
    },
}

const mocks: MockedResponse[] = [
    {
        request: {
            query: GET_INSIGHTS_DASHBOARDS_GQL,
            // variables: { id: undefined },
        },
        result: {
            data: { insightsDashboards: { nodes: [mockDashboard, mockDashboard2] }, currentUser: userMock },
        },
    } as MockedResponse<InsightsDashboardsResult>,
    {
        request: {
            query: GET_INSIGHTS_DASHBOARDS_GQL,
            variables: { id: 'foo' },
        },
        result: {
            data: { insightsDashboards: { nodes: [mockDashboard, mockDashboard2] }, currentUser: userMock },
        },
    } as MockedResponse<InsightsDashboardsResult>,
    {
        request: {
            query: GET_ACCESSIBLE_INSIGHTS_LIST,
        },
        result: {
            data: { insightViews: { nodes: [] } },
        },
    } as MockedResponse<GetAccessibleInsightsListResult>,
    {
        request: {
            query: GET_INSIGHTS_DASHBOARD_OWNERS_GQL,
        },
        result: {
            data: {
                currentUser: userMock,
                site: { id: 'global_instance_id' },
            },
        },
    } as MockedResponse<InsightSubjectsResult>,
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
    } as MockedResponse<GetDashboardInsightsResult>,
    {
        request: {
            query: GET_INSIGHTS_GQL,
            variables: {},
        },
        result: {
            data: { insightViews: { nodes: [] } },
        },
    } as MockedResponse<GetInsightsResult>,
]

const renderDashboardsContent = (
    dashboardID: string = 'foo'
): RenderWithBrandedContextResult & { user: UserEvent } => ({
    user: userEvent,
    ...renderWithBrandedContext(
        <MockedTestProvider mocks={mocks}>
            <Wrapper>
                <DashboardsContentPage dashboardID={dashboardID} telemetryService={mockTelemetryService} />
            </Wrapper>
        </MockedTestProvider>
    ),
})

const triggerDashboardMenuItem = async (screen: RenderWithBrandedContextResult & { user: UserEvent }, name: RegExp) => {
    const { user } = screen
    const dashboardMenu = await waitFor(() => screen.getByRole('button', { name: /Dashboard options/ }))
    user.click(dashboardMenu)

    const dashboardMenuItem = screen.getByRole('menuitem', { name })

    // We're simulating keyboard navigation here to circumvent a bug in ReachUI
    // does not respond to programmatic click events on menu items
    dashboardMenuItem.focus()
    user.keyboard(' ')
}

beforeEach(() => {
    jest.clearAllMocks()
    window.IntersectionObserver = MockIntersectionObserver
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
        const addInsightsButton = await waitFor(() => screen.getByRole('button', { name: /Add or remove insights/ }))

        userEvent.click(addInsightsButton)

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
