import { mount } from 'enzyme'
import * as H from 'history'
import * as React from 'react'
import sinon from 'sinon'

import { CodeMonitorNode } from './CodeMonitoringNode'
import { mockCodeMonitorFields } from './testing/util'

describe('CreateCodeMonitorPage', () => {
    const history = H.createMemoryHistory()

    test('Does not show "Send test email" option when showCodeMonitoringTestEmailButton is false', () => {
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitorFields}
                isSiteAdminUser={true}
                showCodeMonitoringTestEmailButton={false}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(0)
    })

    test('Shows "Send test email" option to site admins on enabled code monitors', () => {
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitorFields}
                isSiteAdminUser={true}
                showCodeMonitoringTestEmailButton={true}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(1)
    })

    test('Does not show "Send test email" option when code monitor is disabled', () => {
        const disabledCodeMonitor = {
            ...mockCodeMonitorFields,
            enabled: false,
        }
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={disabledCodeMonitor}
                isSiteAdminUser={true}
                showCodeMonitoringTestEmailButton={true}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(0)
    })

    test('Does not show "Send test email" option to non-site admins', () => {
        const component = mount(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitorFields}
                isSiteAdminUser={false}
                showCodeMonitoringTestEmailButton={true}
            />
        )
        expect(component.find('.test-send-test-email').length).toBe(0)
    })
})
