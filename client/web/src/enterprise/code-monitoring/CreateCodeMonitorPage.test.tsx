import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router-dom'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { assertAriaDisabled } from '@sourcegraph/testing'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { AuthenticatedUser } from '../../auth'
import type { CreateCodeMonitorVariables } from '../../graphql-operations'

import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'
import { mockCodeMonitor } from './testing/util'

// TODO: these tests trigger an error with CodeMirror, complaining about being
// loaded twice, see https://github.com/uiwjs/react-codemirror/issues/506
describe.skip('CreateCodeMonitorPage', () => {
    const mockUser = {
        id: 'userID',
        username: 'username',
        emails: [{ email: 'user@me.com', isPrimary: true, verified: true }],
        siteAdmin: true,
    } as AuthenticatedUser

    const props = {
        authenticatedUser: mockUser,
        breadcrumbs: [{ depth: 0, breadcrumb: null }],
        setBreadcrumb: sinon.spy(),
        useBreadcrumb: sinon.spy(),
        deleteCodeMonitor: sinon.spy((id: string) => NEVER),
        createCodeMonitor: sinon.spy((monitor: CreateCodeMonitorVariables) =>
            of({ description: mockCodeMonitor.node.description })
        ),
        isLightTheme: true,
        isSourcegraphDotCom: false,
    }

    const origContext = window.context
    beforeEach(() => {
        window.context = {
            emailEnabled: true,
        } as any
    })
    afterEach(() => {
        window.context = origContext
        props.createCodeMonitor.resetHistory()
    })

    test('createCodeMonitor is called on submit', async () => {
        const search = new URLSearchParams({
            'trigger-query': 'test type:diff repo:test',
        }).toString()

        renderWithBrandedContext(
            <MockedTestProvider>
                <Routes>
                    <Route
                        path="/code-monitoring/new"
                        element={<CreateCodeMonitorPage {...props} telemetryRecorder={noOpTelemetryRecorder} />}
                    />
                </Routes>
            </MockedTestProvider>,
            {
                route: '/code-monitoring/new?' + search,
            }
        )
        const nameInput = screen.getByTestId('name-input')
        userEvent.type(nameInput, 'Test updated')

        const triggerInput = screen.getByTestId('trigger-query-edit')
        expect(triggerInput).toBeInTheDocument()
        await waitFor(() => expect(triggerInput).toHaveClass('test-is-valid'))

        userEvent.click(screen.getByTestId('submit-trigger'))
        userEvent.click(screen.getByTestId('form-action-toggle-email'))
        userEvent.click(screen.getByTestId('submit-action-email'))

        userEvent.click(screen.getByTestId('submit-monitor'))
        sinon.assert.called(props.createCodeMonitor)
    })

    test('createCodeMonitor is not called on submit when action is incomplete', () => {
        const search = new URLSearchParams({
            'trigger-query': 'test type:diff repo:test',
        }).toString()

        renderWithBrandedContext(
            <MockedTestProvider>
                <Routes>
                    <Route
                        path="/code-monitoring/new"
                        element={<CreateCodeMonitorPage {...props} telemetryRecorder={noOpTelemetryRecorder} />}
                    />
                </Routes>
            </MockedTestProvider>,
            {
                route: '/code-monitoring/new?' + search,
            }
        )
        const nameInput = screen.getByTestId('name-input')
        userEvent.type(nameInput, 'Test updated')
        userEvent.click(screen.getByTestId('submit-monitor'))

        // Pressing enter does not call createCodeMonitor
        sinon.assert.notCalled(props.createCodeMonitor)

        userEvent.click(screen.getByTestId('form-action-toggle-email'))
        userEvent.click(screen.getByTestId('submit-action-email'))

        // Pressing enter calls createCodeMonitor when all sections are complete
        userEvent.click(screen.getByTestId('submit-monitor'))

        sinon.assert.calledOnce(props.createCodeMonitor)
    })

    test('Actions area button is disabled while trigger is incomplete', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <Routes>
                    <Route
                        path="/code-monitoring/new"
                        element={<CreateCodeMonitorPage {...props} telemetryRecorder={noOpTelemetryRecorder} />}
                    />
                </Routes>
            </MockedTestProvider>,
            { route: '/code-monitoring/new' }
        )
        const actionButton = screen.getByTestId('form-action-toggle-email')
        assertAriaDisabled(actionButton)
    })
})
