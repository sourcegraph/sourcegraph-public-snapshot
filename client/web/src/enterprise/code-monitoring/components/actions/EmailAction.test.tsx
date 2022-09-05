import { MockedResponse } from '@apollo/client/testing'
import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { MonitorEmailPriority, SendTestEmailResult, SendTestEmailVariables } from '../../../../graphql-operations'
import { mockAuthenticatedUser } from '../../testing/util'
import { ActionProps } from '../FormActionArea'

import { EmailAction, SEND_TEST_EMAIL } from './EmailAction'

describe('EmailAction', () => {
    const origContext = window.context
    beforeEach(() => {
        window.context = {
            emailEnabled: true,
        } as any
    })
    afterEach(() => {
        window.context = origContext
    })

    const props: ActionProps = {
        action: undefined,
        setAction: sinon.stub(),
        disabled: false,
        authenticatedUser: mockAuthenticatedUser,
        monitorName: 'Test',
    }

    test('open and submit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <EmailAction {...props} setAction={setActionSpy} />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-email'))
        userEvent.click(getByTestId('submit-action-email'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorEmail',
            enabled: true,
            includeResults: false,
            id: '',
            recipients: { nodes: [{ id: 'userID' }] },
        })
    })

    test('enable include results', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <EmailAction {...props} setAction={setActionSpy} />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-email'))
        userEvent.click(getByTestId('include-results-toggle-email'))
        userEvent.click(getByTestId('submit-action-email'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorEmail',
            enabled: true,
            includeResults: true,
            id: '',
            recipients: { nodes: [{ id: 'userID' }] },
        })
    })

    test('open and delete', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <EmailAction
                    {...props}
                    action={{
                        __typename: 'MonitorEmail',
                        enabled: true,
                        includeResults: false,
                        id: '',
                        recipients: { nodes: [{ id: 'userID' }] },
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-email'))
        userEvent.click(getByTestId('delete-action-email'))

        sinon.assert.calledOnceWithExactly(setActionSpy, undefined)
    })

    test('enable and disable', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <EmailAction
                    {...props}
                    action={{
                        __typename: 'MonitorEmail',
                        enabled: false,
                        includeResults: false,
                        id: '1',
                        recipients: { nodes: [{ id: 'userID' }] },
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        expect(getByTestId('enable-action-toggle-collapsed-email')).not.toBeChecked()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-email'))
        expect(getByTestId('enable-action-toggle-collapsed-email')).toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorEmail',
            enabled: true,
            includeResults: false,
            id: '1',
            recipients: { nodes: [{ id: 'userID' }] },
        })

        setActionSpy.resetHistory()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-email'))
        expect(getByTestId('enable-action-toggle-collapsed-email')).not.toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorEmail',
            enabled: false,
            includeResults: false,
            id: '1',
            recipients: { nodes: [{ id: 'userID' }] },
        })
    })

    test('open, edit, cancel, open again', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <EmailAction
                    {...props}
                    setAction={setActionSpy}
                    action={{
                        __typename: 'MonitorEmail',
                        enabled: true,
                        includeResults: false,
                        id: '1',
                        recipients: { nodes: [{ id: 'userID' }] },
                    }}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-email'))

        expect(getByTestId('enable-action-toggle-expanded-email')).toBeChecked()
        userEvent.click(getByTestId('enable-action-toggle-expanded-email'))
        expect(getByTestId('enable-action-toggle-expanded-email')).not.toBeChecked()
        userEvent.click(getByTestId('cancel-action-email'))

        userEvent.click(getByTestId('form-action-toggle-email'))
        expect(getByTestId('enable-action-toggle-expanded-email')).toBeChecked()

        sinon.assert.notCalled(setActionSpy)
    })

    describe('Send test email', () => {
        const mockedVariables: SendTestEmailVariables = {
            namespace: props.authenticatedUser.id,
            description: props.monitorName,
            email: {
                enabled: true,
                includeResults: false,
                priority: MonitorEmailPriority.NORMAL,
                recipients: [props.authenticatedUser.id],
                header: '',
            },
        }

        test('disabled if no monitor name set', () => {
            const { getByTestId } = render(
                <MockedTestProvider>
                    <EmailAction {...props} monitorName="" />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toBeDisabled()
        })

        test('send test email, success', async () => {
            const mockedResponse: MockedResponse<SendTestEmailResult> = {
                request: {
                    query: SEND_TEST_EMAIL,
                    variables: mockedVariables,
                },
                result: { data: { triggerTestEmailAction: { alwaysNil: null } } },
            }

            const { getByTestId, queryByTestId } = render(
                <MockedTestProvider mocks={[mockedResponse]}>
                    <EmailAction {...props} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            userEvent.click(getByTestId('send-test-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Sending email...')

            await waitForNextApolloResponse()

            expect(getByTestId('send-test-email')).toHaveTextContent('Test email sent!')
            expect(getByTestId('send-test-email')).toBeDisabled()

            expect(queryByTestId('send-test-email-again')).toBeInTheDocument()
            expect(queryByTestId('test-email-error')).not.toBeInTheDocument()
        })

        test('send test email, error', async () => {
            const mockedResponse: MockedResponse<SendTestEmailResult> = {
                request: {
                    query: SEND_TEST_EMAIL,
                    variables: mockedVariables,
                },
                error: new Error('An error occurred'),
            }

            const { getByTestId, queryByTestId } = render(
                <MockedTestProvider mocks={[mockedResponse]}>
                    <EmailAction {...props} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            userEvent.click(getByTestId('send-test-email'))

            await waitForNextApolloResponse()

            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            expect(getByTestId('send-test-email')).toBeEnabled()

            expect(queryByTestId('send-test-email-again')).not.toBeInTheDocument()
            expect(queryByTestId('test-email-error')).toBeInTheDocument()
        })
    })
})
