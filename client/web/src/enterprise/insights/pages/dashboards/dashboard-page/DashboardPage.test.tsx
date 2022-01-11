/* eslint-disable ban/ban */
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createMemoryHistory } from 'history'
import React from 'react'
import { Router, Route } from 'react-router-dom'
import { of } from 'rxjs'
import sinon from 'sinon'

import { CodeInsightsBackend } from '../../../core/backend/code-insights-backend'
import {
    CodeInsightsBackendContext,
    FakeDefaultCodeInsightsBackend,
} from '../../../core/backend/code-insights-backend-context'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../../../core/types/dashboard/virtual-dashboard'

import { DashboardsPage } from './DashboardsPage'
interface ReactRouterMock {
    useHistory: () => unknown
    useRouteMatch: () => unknown
}

const url = '/insights'
const ALL_INSIGHTS = 'All insights'

jest.mock('react-router', () => ({
    ...jest.requireActual<ReactRouterMock>('react-router'),
    useHistory: () => ({
        push: jest.fn(),
    }),
    useRouteMatch: () => ({
        url,
    }),
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

const renderWithRouter = (
    component: React.ReactElement,
    { route = '/', history = createMemoryHistory({ initialEntries: [route] }), api = {} } = {}
) => ({
    ...render(
        <Wrapper api={api}>
            <Router history={history}>
                {component}
                <Route path={`${url}/${ALL_INSIGHTS_DASHBOARD_ID}`}>{ALL_INSIGHTS}</Route>
            </Router>
        </Wrapper>
    ),
    history,
})

describe('DashboardsPage', () => {
    it('should redirect to "All insights" page if no dashboardId is provided', () => {
        const { history } = renderWithRouter(<DashboardsPage telemetryService={mockTelemetryService} />)

        expect(history.location.pathname).toEqual(`${url}/${ALL_INSIGHTS_DASHBOARD_ID}`)
    })

    it('should render dashboard not found page when id is not found', () => {
        renderWithRouter(<DashboardsPage telemetryService={mockTelemetryService} dashboardID="foo" />, {
            api: {
                getDashboardSubjects: () => of([]),
                getDashboards: () => of([]),
            },
        })

        screen.getByText("Hmm, the dashboard wasn't found.")
    })

    it('should log events', () => {
        renderWithRouter(<DashboardsPage telemetryService={mockTelemetryService} dashboardID="foo" />, {
            api: {
                getDashboardSubjects: () => of([]),
                getDashboards: () => of([]),
            },
        })

        expect(mockTelemetryService.logViewEvent.calledWith('Insights')).toBe(true)

        userEvent.click(screen.getByText('Create new insight'))
        expect(mockTelemetryService.log.calledWith('InsightAddMoreClick')).toBe(true)
    })
})
