import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import * as H from 'history'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import {
    MonitorEditInput,
    MonitorEditTriggerInput,
    MonitorEditActionInput,
    MonitorEmailPriority,
} from '@sourcegraph/shared/src/graphql-operations'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { FetchCodeMonitorResult } from '../../graphql-operations'

import { ManageCodeMonitorPage } from './ManageCodeMonitorPage'
import { mockCodeMonitor, mockCodeMonitorFields, mockUser } from './testing/util'

describe('ManageCodeMonitorPage', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = {
            emailEnabled: true,
        } as any
    })
    afterEach(() => {
        window.context = origContext
    })

    const history = H.createMemoryHistory()
    history.location.pathname = '/code-monitoring/test-monitor-id'
    const props = {
        history,
        location: history.location,
        authenticatedUser: mockUser,
        breadcrumbs: [{ depth: 0, breadcrumb: null }],
        setBreadcrumb: sinon.spy(),
        useBreadcrumb: sinon.spy(),
        fetchUserCodeMonitors: sinon.spy(),
        updateCodeMonitor: sinon.spy(
            (
                monitorEditInput: MonitorEditInput,
                triggerEditInput: MonitorEditTriggerInput,
                actionEditInput: MonitorEditActionInput[]
            ) => of(mockCodeMonitorFields)
        ),
        fetchCodeMonitor: sinon.spy((id: string) => of(mockCodeMonitor as FetchCodeMonitorResult)),
        match: {
            params: { id: 'test-id' },
            isExact: true,
            path: history.location.pathname,
            url: 'https://sourcegraph.com',
        },
        toggleCodeMonitorEnabled: sinon.spy((id: string, enabled: boolean) => of({ id: 'test', enabled: true })),
        deleteCodeMonitor: sinon.spy((id: string) => NEVER),
        isLightTheme: false,
        isSourcegraphDotCom: false,
    }

    test('Form is pre-loaded with code monitor data', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        expect(props.fetchCodeMonitor.calledOnce).toBe(true)

        const nameInput = screen.getByTestId('name-input')
        expect(nameInput).toBeInTheDocument()
        expect(nameInput).toHaveValue('Test code monitor')

        const currentQueryValue = screen.getByTestId('trigger-query-existing')
        const currentActionEmailValue = screen.getByTestId('existing-action-email')
        expect(currentQueryValue).toHaveTextContent('test')
        expect(currentActionEmailValue).toHaveTextContent('user@me.com')
        props.fetchCodeMonitor.resetHistory()
    })

    test('Updating the form executes the update request', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const nameInput = screen.getByTestId('name-input')
        expect(nameInput).toHaveValue('Test code monitor')

        userEvent.clear(nameInput)
        userEvent.type(nameInput, 'Test updated')
        const submitButton = screen.getByTestId('submit-monitor')
        userEvent.click(submitButton)
        sinon.assert.called(props.updateCodeMonitor)
        sinon.assert.calledWith(
            props.updateCodeMonitor,
            {
                id: 'test-id',
                update: { namespace: 'userID', description: 'Test updated', enabled: true },
            },
            { id: 'test-0', update: { query: 'test' } },
            [
                {
                    email: {
                        id: 'test-action-0',
                        update: {
                            enabled: true,
                            includeResults: false,
                            priority: MonitorEmailPriority.NORMAL,
                            recipients: ['userID'],
                            header: '',
                        },
                    },
                },
                {
                    slackWebhook: {
                        id: 'test-action-1',
                        update: {
                            enabled: true,
                            includeResults: false,
                            url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX',
                        },
                    },
                },
            ]
        )
        props.updateCodeMonitor.resetHistory()
    })

    test('Clicking Edit in the trigger area opens the query form', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        expect(screen.queryByTestId('trigger-query-edit')).not.toBeInTheDocument()
        userEvent.click(screen.getByTestId('trigger-button'))
        expect(screen.getByTestId('trigger-query-edit')).toBeInTheDocument()
    })

    test('Clicking Edit in the action area opens the action form', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        expect(screen.queryByTestId('action-form-email')).not.toBeInTheDocument()
        const editTrigger = screen.getByTestId('form-action-toggle-email')
        userEvent.click(editTrigger)
        expect(screen.queryByTestId('action-form-email')).toBeInTheDocument()
    })

    test('Save button is disabled when no changes have been made, enabled when changes have been made', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const submitButton = screen.getByTestId('submit-monitor')
        expect(submitButton).toBeDisabled()

        userEvent.type(screen.getByTestId('name-input'), 'Test code monitor updated')

        expect(submitButton).toBeEnabled()
    })

    test('Cancelling after changes have been made shows confirmation prompt', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const confirmStub = sinon.stub(window, 'confirm')

        userEvent.type(screen.getByTestId('name-input'), 'Test code monitor updated')
        userEvent.click(screen.getByTestId('cancel-monitor'))

        sinon.assert.calledOnce(confirmStub)
        confirmStub.restore()
    })

    test('Cancelling without any changes made does not show confirmation prompt', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        const confirmStub = sinon.stub(window, 'confirm')
        userEvent.click(screen.getByTestId('cancel-monitor'))

        sinon.assert.notCalled(confirmStub)
        confirmStub.restore()
    })

    test('Clicking delete code monitor opens deletion confirmation modal', () => {
        renderWithBrandedContext(
            <MockedTestProvider>
                <ManageCodeMonitorPage {...props} />
            </MockedTestProvider>
        )
        userEvent.click(screen.getByTestId('delete-monitor'))
        expect(screen.getByTestId('delete-modal')).toBeInTheDocument()

        const confirmDeleteButton = screen.getByTestId('confirm-delete-monitor')
        expect(confirmDeleteButton).toBeInTheDocument()
        userEvent.click(confirmDeleteButton)

        sinon.assert.calledOnce(props.deleteCodeMonitor)
        expect(props.history.location.pathname).toEqual('/code-monitoring')
    })
})
