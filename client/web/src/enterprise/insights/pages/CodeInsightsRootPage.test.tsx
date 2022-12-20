/* eslint-disable ban/ban */
import React from 'react'

import { MockedResponse } from '@apollo/client/testing'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import * as H from 'history'
import { MemoryRouter } from 'react-router'
import { Route } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'
import sinon from 'sinon'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/testing/MockIntersectionObserver'

import { InsightsDashboardsResult } from '../../../graphql-operations'
import { CodeInsightsBackend, CodeInsightsBackendContext, FakeDefaultCodeInsightsBackend } from '../core'
import { GET_INSIGHT_DASHBOARDS_GQL } from '../core/hooks/use-insight-dashboards'

import { CodeInsightsRootPage, CodeInsightsRootPageTab } from './CodeInsightsRootPage'

interface ReactRouterMock {
    useHistory: () => unknown
    useRouteMatch: () => unknown
}

const url = '/insights'

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
    logPageView: sinon.spy(),
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

const renderWithBrandedContext = (component: React.ReactElement, { route = '/', api = {} } = {}) => {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const routerSettings: { testHistory: H.History; testLocation: H.Location } = {}

    return {
        ...render(
            <MockedTestProvider mocks={mockedGQL}>
                <Wrapper api={api}>
                    <MemoryRouter initialEntries={[route]}>
                        <CompatRouter>
                            {component}
                            <Route
                                path="*"
                                render={({ history, location }) => {
                                    routerSettings.testHistory = history
                                    routerSettings.testLocation = location
                                    return null
                                }}
                            />
                        </CompatRouter>
                    </MemoryRouter>
                </Wrapper>
            </MockedTestProvider>
        ),
        ...routerSettings,
    }
}

describe('CodeInsightsRootPage', () => {
    beforeAll(() => {
        window.IntersectionObserver = MockIntersectionObserver
    })

    it('should render dashboard not found page when id is not found', () => {
        renderWithBrandedContext(
            <CodeInsightsRootPage
                activeTab={CodeInsightsRootPageTab.Dashboards}
                telemetryService={mockTelemetryService}
            />,
            {
                route: '/insights/dashboards/foo',
            }
        )

        screen.findByText("Hmm, the dashboard wasn't found.")
    })

    it('should log events', () => {
        renderWithBrandedContext(
            <CodeInsightsRootPage
                activeTab={CodeInsightsRootPageTab.Dashboards}
                telemetryService={mockTelemetryService}
            />,
            {
                route: '/insights/dashboards/foo',
            }
        )

        userEvent.click(screen.getByText('Create insight'))
        expect(mockTelemetryService.log.calledWith('InsightAddMoreClick')).toBe(true)
    })
})
