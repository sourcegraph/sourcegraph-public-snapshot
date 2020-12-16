import { CodeMonitorNode } from './CodeMonitoringNode'
import * as React from 'react'
import * as H from 'history'
import { AuthenticatedUser } from '../../auth'
import sinon from 'sinon'
import { mount } from 'enzyme'
import { mockCodeMonitor } from './testing/util'

describe('CreateCodeMonitorPage', () => {
    const mockUser = {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser

    const history = H.createMemoryHistory()

    test('Shows "Send test email" option to site admins on enabled code monitors', () => {
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitor.node}
                authentictedUser={mockUser}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(1)
    })

    test('Does not show "Send test email" option when code monitor is disabled', () => {
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={{ ...mockCodeMonitor.node, enabled: false }}
                authentictedUser={mockUser}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(0)
    })

    test('Does not show "Send test email" option to non-site admins', () => {
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitor.node}
                authentictedUser={{ ...mockUser, siteAdmin: false }}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(0)
    })
})
