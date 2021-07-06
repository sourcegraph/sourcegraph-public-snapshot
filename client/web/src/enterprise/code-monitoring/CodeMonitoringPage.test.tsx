import { mount } from 'enzyme'
import * as H from 'history'
import * as React from 'react'
import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../auth'
import { EMPTY_FEATURE_FLAGS } from '../../featureFlags/featureFlags'
import { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitoringPage } from './CodeMonitoringPage'
import { mockCodeMonitorNodes } from './testing/util'

const additionalProps = {
    authenticatedUser: { id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser,
    fetchUserCodeMonitors: ({ id, first, after }: ListUserCodeMonitorsVariables) =>
        of({
            nodes: mockCodeMonitorNodes,
            pageInfo: {
                endCursor: 'foo10',
                hasNextPage: true,
            },
            totalCount: 12,
        }),
    toggleCodeMonitorEnabled: sinon.spy((id: string, enabled: boolean) => of({ id: 'test', enabled: true })),
    settingsCascade: EMPTY_SETTINGS_CASCADE,
    isLightTheme: false,
    featureFlags: EMPTY_FEATURE_FLAGS,
}

const generateMockFetchMonitors = (count: number) => ({ id, first, after }: ListUserCodeMonitorsVariables) => {
    const result: ListCodeMonitors = {
        nodes: mockCodeMonitorNodes.slice(0, count),
        pageInfo: {
            endCursor: `foo${count}`,
            hasNextPage: count > 10,
        },
        totalCount: count,
    }
    return of(result)
}

describe('CodeMonitoringListPage', () => {
    test('Code monitoring page with less than 10 results', () => {
        expect(
            mount(
                <MemoryRouter initialEntries={['/code-monitoring']}>
                    <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(3)} />
                </MemoryRouter>
            )
        ).toMatchSnapshot()
    })
    test('Code monitoring page with 10 results', () => {
        expect(
            mount(
                <MemoryRouter initialEntries={['/code-monitoring']}>
                    <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(10)} />
                </MemoryRouter>
            )
        ).toMatchSnapshot()
    })
    test('Code monitoring page with more than 10 results', () => {
        expect(
            mount(
                <MemoryRouter initialEntries={['/code-monitoring']}>
                    <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(12)} />
                </MemoryRouter>
            )
        ).toMatchSnapshot()
    })

    test('Clicking enabled toggle calls toggleCodeMonitorEnabled', () => {
        const component = mount(
            <MemoryRouter initialEntries={['/code-monitoring']}>
                <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(1)} />
            </MemoryRouter>
        )
        const toggle = component.find('.test-toggle-monitor-enabled')
        toggle.simulate('click')
        expect(additionalProps.toggleCodeMonitorEnabled.calledOnce)
    })

    test('Redirect to getting started if empty', () => {
        const component = mount(
            <MemoryRouter initialEntries={['/code-monitoring']}>
                <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(0)} />
            </MemoryRouter>
        )

        const history: H.History = component.find('Router').prop('history')
        expect(history.location.pathname).toBe('/code-monitoring/getting-started')
    })

    test('Do not redirect to getting started if not empty', () => {
        const component = mount(
            <MemoryRouter initialEntries={['/code-monitoring']}>
                <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(1)} />
            </MemoryRouter>
        )

        const history: H.History = component.find('Router').prop('history')
        expect(history.location.pathname).toBe('/code-monitoring')
    })

    test('Redirect to sign in if not logged in', () => {
        const component = mount(
            <MemoryRouter initialEntries={['/code-monitoring']}>
                <CodeMonitoringPage {...additionalProps} authenticatedUser={null} />
            </MemoryRouter>
        )

        const history: H.History = component.find('Router').prop('history')
        expect(history.location.pathname).toBe('/sign-in')
    })

    test('Redirect to getting started if not logged in and feature flag is enabled', () => {
        const component = mount(
            <MemoryRouter initialEntries={['/code-monitoring']}>
                <CodeMonitoringPage
                    {...additionalProps}
                    authenticatedUser={null}
                    featureFlags={new Map([['w1-signup-optimisation', true]])}
                />
            </MemoryRouter>
        )

        const history: H.History = component.find('Router').prop('history')
        expect(history.location.pathname).toBe('/code-monitoring/getting-started')
    })
})
