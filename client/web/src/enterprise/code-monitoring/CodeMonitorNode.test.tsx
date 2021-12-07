import { render, screen } from '@testing-library/react'
import * as H from 'history'
import * as React from 'react'
import sinon from 'sinon'

import { CodeMonitorNode } from './CodeMonitoringNode'
import { mockCodeMonitorFields } from './testing/util'

describe('CreateCodeMonitorPage', () => {
    const history = H.createMemoryHistory()

    test('Does not show "Send test email" option when showCodeMonitoringTestEmailButton is false', () => {
        render(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitorFields}
                isSiteAdminUser={true}
                showCodeMonitoringTestEmailButton={false}
            />
        )
        expect(screen.queryByRole('button', { name: /Send test email/ })).not.toBeInTheDocument()
    })

    test('Shows "Send test email" option to site admins on enabled code monitors', () => {
        render(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitorFields}
                isSiteAdminUser={true}
                showCodeMonitoringTestEmailButton={true}
            />
        )
        expect(screen.getByRole('button', { name: /Send test email/ })).toBeInTheDocument()
    })

    test('Does not show "Send test email" option when code monitor is disabled', () => {
        const disabledCodeMonitor = {
            ...mockCodeMonitorFields,
            enabled: false,
        }
        render(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={disabledCodeMonitor}
                isSiteAdminUser={true}
                showCodeMonitoringTestEmailButton={true}
            />
        )
        expect(screen.queryByRole('button', { name: /Send test email/ })).not.toBeInTheDocument()
    })

    test('Does not show "Send test email" option to non-site admins', () => {
        render(
            <CodeMonitorNode
                toggleCodeMonitorEnabled={sinon.spy()}
                location={history.location}
                node={mockCodeMonitorFields}
                isSiteAdminUser={false}
                showCodeMonitoringTestEmailButton={true}
            />
        )
        expect(screen.queryByRole('button', { name: /Send test email/ })).not.toBeInTheDocument()
    })
})
