import { mount } from 'enzyme'
import * as React from 'react'
import * as H from 'history'
import { CodeMonitoringPage } from './CodeMonitoringPage'
import { of } from 'rxjs'
import { AuthenticatedUser } from '../../auth'
import { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'
import sinon from 'sinon'
import { mockCodeMonitorNodes } from './testing/util'
import { EMPTY_SETTINGS_CASCADE } from '../../../../shared/src/settings/settings'

const history = H.createBrowserHistory({})
history.replace({ pathname: '/code-monitoring' })
const additionalProps = {
    history,
    location: history.location,
    authenticatedUser: { id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser,
    breadcrumbs: [{ depth: 0, breadcrumb: null }],
    setBreadcrumb: sinon.spy(),
    useBreadcrumb: sinon.spy(),
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
            mount(<CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(3)} />)
        ).toMatchSnapshot()
    })
    test('Code monitoring page with 10 results', () => {
        expect(
            mount(<CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(10)} />)
        ).toMatchSnapshot()
    })
    test('Code monitoring page with more than 10 results', () => {
        expect(
            mount(<CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(12)} />)
        ).toMatchSnapshot()
    })

    test('Clicking enabled toggle calls toggleCodeMonitorEnabled', () => {
        const component = mount(
            <CodeMonitoringPage {...additionalProps} fetchUserCodeMonitors={generateMockFetchMonitors(1)} />
        )
        const toggle = component.find('.test-toggle-monitor-enabled')
        toggle.simulate('click')
        expect(additionalProps.toggleCodeMonitorEnabled.calledOnce)
    })
})
