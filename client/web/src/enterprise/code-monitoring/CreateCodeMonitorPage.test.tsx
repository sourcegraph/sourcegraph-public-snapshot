import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import * as H from 'history'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { assertAriaDisabled } from '@sourcegraph/shared/dev/aria-asserts'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'

import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'
import { mockCodeMonitor } from './testing/util'

describe('CreateCodeMonitorPage', () => {
    const mockUser = {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser

    const history = H.createMemoryHistory()
    const props = {
        location: history.location,
        authenticatedUser: mockUser,
        breadcrumbs: [{ depth: 0, breadcrumb: null }],
        setBreadcrumb: sinon.spy(),
        useBreadcrumb: sinon.spy(),
        history,
        deleteCodeMonitor: sinon.spy(() => NEVER),
        createCodeMonitor: sinon.spy(() => of({ description: mockCodeMonitor.node.description })),
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
                <CreateCodeMonitorPage {...props} location={{ ...history.location, search }} />
            </MockedTestProvider>
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
                <CreateCodeMonitorPage {...props} location={{ ...history.location, search }} />
            </MockedTestProvider>
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
                <CreateCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const actionButton = screen.getByTestId('form-action-toggle-email')
        assertAriaDisabled(actionButton)
    })
})
