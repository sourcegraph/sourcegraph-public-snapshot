import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import * as H from 'history'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { AuthenticatedUser } from '../../auth'
import { CreateCodeMonitorVariables } from '../../graphql-operations'

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
        deleteCodeMonitor: sinon.spy((id: string) => NEVER),
        createCodeMonitor: sinon.spy((monitor: CreateCodeMonitorVariables) =>
            of({ description: mockCodeMonitor.node.description })
        ),
        isLightTheme: true,
        isSourcegraphDotCom: false,
    }

    afterEach(() => {
        props.createCodeMonitor.resetHistory()
    })

    test('createCodeMonitor is called on submit', async () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <CreateCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const nameInput = screen.getByTestId('name-input')
        userEvent.type(nameInput, 'Test updated')
        userEvent.click(screen.getByTestId('trigger-button'))

        const triggerInput = screen.getByTestId('trigger-query-edit')
        expect(triggerInput).toBeInTheDocument()

        await waitFor(() => expect(triggerInput.querySelector('textarea[role="textbox"]')).toBeInTheDocument())

        const textbox = triggerInput.querySelector('textarea[role="textbox"]')!
        userEvent.type(textbox, 'test type:diff repo:test')

        await waitFor(() => expect(triggerInput).toHaveClass('test-is-valid'))

        userEvent.click(screen.getByTestId('submit-trigger'))
        userEvent.click(screen.getByTestId('form-action-toggle-email'))
        userEvent.click(screen.getByTestId('submit-action-email'))

        userEvent.click(screen.getByTestId('submit-monitor'))
        sinon.assert.called(props.createCodeMonitor)
    })

    test('createCodeMonitor is not called on submit when trigger or action is incomplete', async () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <CreateCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const nameInput = screen.getByTestId('name-input')
        userEvent.type(nameInput, 'Test updated')
        userEvent.click(screen.getByTestId('submit-monitor'))

        // Pressing enter does not call createCodeMonitor because other fields not complete
        sinon.assert.notCalled(props.createCodeMonitor)

        userEvent.click(screen.getByTestId('trigger-button'))

        const triggerInput = screen.getByTestId('trigger-query-edit')
        expect(triggerInput).toBeInTheDocument()

        await waitFor(() => expect(triggerInput.querySelector('textarea[role="textbox"]')).toBeInTheDocument())

        const textbox = triggerInput.querySelector('textarea[role="textbox"]')!
        userEvent.type(textbox, 'test type:diff repo:test')

        await waitFor(() => expect(triggerInput).toHaveClass('test-is-valid'))

        userEvent.click(screen.getByTestId('submit-trigger'))
        userEvent.click(screen.getByTestId('submit-monitor'))

        // Pressing enter still does not call createCodeMonitor
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
        expect(actionButton).toBeDisabled()
    })
})
