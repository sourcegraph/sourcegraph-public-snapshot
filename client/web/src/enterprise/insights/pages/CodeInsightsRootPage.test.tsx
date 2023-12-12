/* eslint-disable ban/ban */
import type { ReactElement } from 'react'

import type { MockedResponse } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import sinon from 'sinon'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/testing/MockIntersectionObserver'

import type { InsightsDashboardsResult } from '../../../graphql-operations'
import { type CodeInsightsBackend, CodeInsightsBackendContext, FakeDefaultCodeInsightsBackend } from '../core'
import { GET_INSIGHT_DASHBOARDS_GQL } from '../core/hooks/use-insight-dashboards'

import { CodeInsightsRootPage, CodeInsightsRootPageTab } from './CodeInsightsRootPage'

function mockRouterDom() {
    return {
        ...jest.requireActual('react-router-dom'),
        useNavigate: () => ({
            push: jest.fn(),
        }),
    }
}

jest.mock('react-router-dom', () => mockRouterDom())

const mockTelemetryService = {
    log: sinon.spy(),
    logViewEvent: sinon.spy(),
    logPageView: sinon.spy(),
}

const mockTelemetryRecorder = {
    log: sinon.spy(),
    recordEvent: sinon.spy(),
    recordPageView: sinon.spy(),
}

const fakeApi = new FakeDefaultCodeInsightsBackend()

const Wrapper: React.FunctionComponent<React.PropsWithChildren<{ api: Partial<CodeInsightsBackend> }>> = ({
    children,
    api = {},
}) => {
    const extendedApi: CodeInsightsBackend = {
        ...fakeApi,
        ...api,
    }
    return <CodeInsightsBackendContext.Provider value={extendedApi}>{children}</CodeInsightsBackendContext.Provider>
}

const mockedGQL: MockedResponse[] = [
    {
        request: {
            query: getDocumentNode(GET_INSIGHT_DASHBOARDS_GQL),
        },
        result: {
            data: {
                insightsDashboards: {
                    nodes: [
                        {
                            __typename: 'InsightsDashboard',
                            id: 'foo',
                            title: 'Global Dashboard',
                            grants: {
                                __typename: 'InsightsPermissionGrants',
                                users: [],
                                organizations: [],
                                global: true,
                            },
                        },
                    ],
                },
                currentUser: {
                    __typename: 'User',
                    id: '001',
                    organizations: {
                        nodes: [],
                    },
                },
            },
        },
    } as MockedResponse<InsightsDashboardsResult>,
]

const renderWithBrandedContext = (component: ReactElement, { route = '/', path = '*', api = {} } = {}) => ({
    ...render(
        <MockedTestProvider mocks={mockedGQL}>
            <Wrapper api={api}>
                <MemoryRouter initialEntries={[route]}>
                    <Routes>
                        <Route path={path} element={component} />
                    </Routes>
                </MemoryRouter>
            </Wrapper>
        </MockedTestProvider>
    ),
})

describe('CodeInsightsRootPage', () => {
    beforeAll(() => {
        window.IntersectionObserver = MockIntersectionObserver
    })

    it('should render dashboard not found page when id is not found', () => {
        renderWithBrandedContext(
            <CodeInsightsRootPage
                activeTab={CodeInsightsRootPageTab.Dashboards}
                telemetryService={mockTelemetryService}
                telemetryRecorder={mockTelemetryRecorder}
            />,
            {
                route: '/insights/dashboards/foo',
                path: '/insights/dashboards/:dashboardId',
            }
        )

        screen.findByText("Hmm, the dashboard wasn't found.")
    })

    it('should log events', () => {
        renderWithBrandedContext(
            <CodeInsightsRootPage
                activeTab={CodeInsightsRootPageTab.Dashboards}
                telemetryService={mockTelemetryService}
                telemetryRecorder={mockTelemetryRecorder}
            />,
            {
                route: '/insights/dashboards/foo',
                path: '/insights/dashboards/:dashboardId',
            }
        )

        userEvent.click(screen.getByText('Create insight'))
        expect(mockTelemetryService.log.calledWith('InsightAddMoreClick')).toBe(true)
        expect(mockTelemetryRecorder.log.calledWith('InsightAddMoreClick')).toBe(true)
    })
})
