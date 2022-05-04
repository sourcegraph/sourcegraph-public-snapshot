import { MockedResponse } from '@apollo/client/testing'
import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { SendTestSlackWebhookResult, SendTestSlackWebhookVariables } from '../../../../graphql-operations'
import { mockAuthenticatedUser } from '../../testing/util'
import { ActionProps, MonitorAction } from '../FormActionArea'

import { SEND_TEST_SLACK_WEBHOOK, SlackWebhookAction } from './SlackWebhookAction'

const SLACK_URL = 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX'

describe('SlackWebhookAction', () => {
    const props: ActionProps = {
        action: undefined,
        setAction: sinon.stub(),
        disabled: false,
        monitorName: 'Test',
        authenticatedUser: mockAuthenticatedUser,
    }

    test('open and submit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <SlackWebhookAction {...props} setAction={setActionSpy} />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))

        expect(getByTestId('submit-action-slack-webhook')).toBeDisabled()

        userEvent.type(getByTestId('slack-webhook-url'), SLACK_URL)
        expect(getByTestId('submit-action-slack-webhook')).toBeEnabled()

        userEvent.click(getByTestId('include-results-toggle-slack-webhook'))

        userEvent.click(getByTestId('submit-action-slack-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: true,
            includeResults: true,
            id: '',
            url: SLACK_URL,
        })
    })

    test('open and edit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <SlackWebhookAction
                    {...props}
                    setAction={setActionSpy}
                    action={{
                        __typename: 'MonitorSlackWebhook',
                        enabled: true,
                        includeResults: false,
                        id: '1',
                        url: SLACK_URL,
                    }}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
        expect(getByTestId('submit-action-slack-webhook')).toBeEnabled()

        userEvent.clear(getByTestId('slack-webhook-url'))
        expect(getByTestId('submit-action-slack-webhook')).toBeDisabled()

        userEvent.type(getByTestId('slack-webhook-url'), SLACK_URL)
        expect(getByTestId('submit-action-slack-webhook')).toBeEnabled()

        userEvent.click(getByTestId('submit-action-slack-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: true,
            includeResults: false,
            id: '1',
            url: SLACK_URL,
        })
    })

    test('open and delete', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <SlackWebhookAction
                    {...props}
                    action={{
                        __typename: 'MonitorSlackWebhook',
                        enabled: true,
                        includeResults: false,
                        id: '2',
                        url: SLACK_URL,
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
        userEvent.click(getByTestId('delete-action-slack-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, undefined)
    })

    test('enable and disable', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <SlackWebhookAction
                    {...props}
                    action={{
                        __typename: 'MonitorSlackWebhook',
                        enabled: false,
                        includeResults: false,
                        id: '5',
                        url: SLACK_URL,
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        expect(getByTestId('enable-action-toggle-collapsed-slack-webhook')).not.toBeChecked()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-slack-webhook'))
        expect(getByTestId('enable-action-toggle-collapsed-slack-webhook')).toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: true,
            includeResults: false,
            id: '5',
            url: SLACK_URL,
        })

        setActionSpy.resetHistory()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-slack-webhook'))
        expect(getByTestId('enable-action-toggle-collapsed-slack-webhook')).not.toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: false,
            includeResults: false,
            id: '5',
            url: SLACK_URL,
        })
    })

    test('open, edit, cancel, open again', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <MockedTestProvider>
                <SlackWebhookAction
                    {...props}
                    action={{
                        __typename: 'MonitorSlackWebhook',
                        enabled: true,
                        includeResults: false,
                        id: '5',
                        url: 'https://example.com',
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))

        expect(getByTestId('enable-action-toggle-expanded-slack-webhook')).toBeChecked()
        userEvent.click(getByTestId('enable-action-toggle-expanded-slack-webhook'))
        expect(getByTestId('enable-action-toggle-expanded-slack-webhook')).not.toBeChecked()

        userEvent.type(getByTestId('slack-webhook-url'), 'https://example2.com')

        userEvent.click(getByTestId('cancel-action-slack-webhook'))

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
        expect(getByTestId('slack-webhook-url')).toHaveValue('https://example.com')
        expect(getByTestId('enable-action-toggle-expanded-slack-webhook')).toBeChecked()

        sinon.assert.notCalled(setActionSpy)
    })

    describe('Send test message', () => {
        const mockAction: MonitorAction = {
            __typename: 'MonitorSlackWebhook',
            enabled: false,
            includeResults: false,
            id: '5',
            url: SLACK_URL,
        }

        const mockedVars: SendTestSlackWebhookVariables = {
            namespace: props.authenticatedUser.id,
            description: props.monitorName,
            slackWebhook: {
                enabled: true,
                includeResults: false,
                url: mockAction.url,
            },
        }

        test('disabled if no webhook url set', () => {
            const { getByTestId } = render(
                <MockedTestProvider>
                    <SlackWebhookAction {...props} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
            expect(getByTestId('send-test-slack-webhook')).toBeDisabled()
        })

        test('disabled if no monitor name set', () => {
            const { getByTestId } = render(
                <MockedTestProvider>
                    <SlackWebhookAction {...props} monitorName="" />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
            expect(getByTestId('send-test-slack-webhook')).toBeDisabled()
        })

        test('send test message, success', async () => {
            const mockedResponse: MockedResponse<SendTestSlackWebhookResult> = {
                request: {
                    query: SEND_TEST_SLACK_WEBHOOK,
                    variables: mockedVars,
                },
                result: { data: { triggerTestSlackWebhookAction: { alwaysNil: null } } },
            }

            const { getByTestId, queryByTestId } = render(
                <MockedTestProvider mocks={[mockedResponse]}>
                    <SlackWebhookAction {...props} action={mockAction} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
            expect(getByTestId('send-test-slack-webhook')).toHaveTextContent('Send test message')

            userEvent.click(getByTestId('send-test-slack-webhook'))
            expect(getByTestId('send-test-slack-webhook')).toHaveTextContent('Sending message...')

            await waitForNextApolloResponse()

            expect(getByTestId('send-test-slack-webhook')).toHaveTextContent('Test message sent!')
            expect(getByTestId('send-test-slack-webhook')).toBeDisabled()

            expect(queryByTestId('send-test-slack-webhook')).toBeInTheDocument()
            expect(queryByTestId('test-email-slack-webhook')).not.toBeInTheDocument()
        })

        test('send test message, error', async () => {
            const mockedResponse: MockedResponse<SendTestSlackWebhookResult> = {
                request: {
                    query: SEND_TEST_SLACK_WEBHOOK,
                    variables: mockedVars,
                },
                error: new Error('An error occurred'),
            }

            const { getByTestId, queryByTestId } = render(
                <MockedTestProvider mocks={[mockedResponse]}>
                    <SlackWebhookAction {...props} action={mockAction} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
            expect(getByTestId('send-test-slack-webhook')).toHaveTextContent('Send test message')

            userEvent.click(getByTestId('send-test-slack-webhook'))

            await waitForNextApolloResponse()

            expect(getByTestId('send-test-slack-webhook')).toHaveTextContent('Send test message')

            expect(getByTestId('send-test-slack-webhook')).toBeEnabled()

            expect(queryByTestId('send-test-slack-webhook-again')).not.toBeInTheDocument()
            expect(queryByTestId('test-slack-webhook-error')).toBeInTheDocument()
        })
    })
})
