import { mount } from 'enzyme'
import * as React from 'react'
import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../auth'
import { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitoringPage } from './CodeMonitoringPage'
import { mockCodeMonitorNodes } from './testing/util'

jest.mock('../../tracking/eventLogger', () => ({
    eventLogger: { logViewEvent: () => undefined },
}))

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
})
