import userEvent from '@testing-library/user-event'
import React from 'react'
import { of } from 'rxjs'
import sinon from 'sinon'

import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import { AuthenticatedUser } from '../../../../../../../auth'
import { CodeInsightsBackend } from '../../../../../core/backend/code-insights-backend'
import {
    CodeInsightsBackendContext,
    FakeDefaultCodeInsightsBackend,
} from '../../../../../core/backend/code-insights-backend-context'
import { InsightDashboard, InsightsDashboardScope, InsightsDashboardType } from '../../../../../core/types'

import { DashboardsContent } from './DashboardsContent'

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

const fakeApi = new FakeDefaultCodeInsightsBackend()

const Wrapper: React.FunctionComponent<{ api: Partial<CodeInsightsBackend> }> = ({ children, api = {} }) => {
    const extendedApi: CodeInsightsBackend = {
        ...fakeApi,
        ...api,
    }
    return <CodeInsightsBackendContext.Provider value={extendedApi}>{children}</CodeInsightsBackendContext.Provider>
}

const mockDashboard: InsightDashboard = {
    id: 'foo',
    type: InsightsDashboardType.Custom,
    title: 'Global Dashboard',
    settingsKey: null,
    scope: InsightsDashboardScope.Global,
}

const mockDashboard2: InsightDashboard = {
    id: 'bar',
    type: InsightsDashboardType.Custom,
    title: 'Global Dashboard 2',
    settingsKey: null,
    scope: InsightsDashboardScope.Global,
}

const renderDashboardsContent = (component: React.ReactElement, { api = {} } = {}): RenderWithRouterResult => {
    const mergedApi = {
        getDashboardSubjects: () => of([]),
        getInsights: () => of([]),
        getDashboards: () => of([mockDashboard, mockDashboard2]),
        ...api,
    }
    return renderWithRouter(<Wrapper api={mergedApi}>{component}</Wrapper>)
}

beforeEach(() => {
    jest.clearAllMocks()
})

describe('DashboardsContent', () => {
    it('renders dashboard not found', () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="foo" telemetryService={mockTelemetryService} />,
            {
                api: {
                    getDashboards: () => of([]),
                },
            }
        )

        screen.getByText("Hmm, the dashboard wasn't found.")
    })

    it('renders a dashboard', () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="foo" telemetryService={mockTelemetryService} />
        )

        screen.getByRole('button', { name: /Global Dashboard/ })
    })

    it('redirect to new dashboard page on selection', () => {
        const screen = renderDashboardsContent(
            <DashboardsContent dashboardID="foo" telemetryService={mockTelemetryService} />
        )
        const { history } = screen

        const chooseDashboard = screen.getByRole('button', { name: /Choose a dashboard/ })
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
