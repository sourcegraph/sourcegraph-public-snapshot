import type { MockedResponse } from '@apollo/client/testing'
import { describe, expect, test } from '@jest/globals'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { assertAriaDisabled, assertAriaEnabled } from '@sourcegraph/testing'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { SendTestWebhookResult, SendTestWebhookVariables } from '../../../../graphql-operations'
import { mockAuthenticatedUser } from '../../testing/util'
import type { ActionProps, MonitorAction } from '../FormActionArea'

import { SEND_TEST_WEBHOOK, WebhookAction } from './WebhookAction'

describe('WebhookAction', () => {
    const props: ActionProps = {
        action: undefined,
        setAction: sinon.stub(),
        disabled: false,
        monitorName: 'Test',
        authenticatedUser: mockAuthenticatedUser,
    }

    test('open and submit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <WebhookAction {...props} setAction={setActionSpy} />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-webhook'))

        assertAriaDisabled(getByTestId('submit-action-webhook'))

        userEvent.type(getByTestId('webhook-url'), 'https://example.com')
        assertAriaEnabled(getByTestId('submit-action-webhook'))

        userEvent.click(getByTestId('include-results-toggle-webhook'))

        userEvent.click(getByTestId('submit-action-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorWebhook',
            enabled: true,
            includeResults: true,
            id: '',
            url: 'https://example.com',
        })
    })

    test('open and edit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <WebhookAction
                    {...props}
                    setAction={setActionSpy}
                    action={{
                        __typename: 'MonitorWebhook',
                        enabled: true,
                        includeResults: false,
                        id: '1',
                        url: 'https://example.com',
                    }}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-webhook'))
        assertAriaEnabled(getByTestId('submit-action-webhook'))

        userEvent.clear(getByTestId('webhook-url'))
        assertAriaDisabled(getByTestId('submit-action-webhook'))

        userEvent.type(getByTestId('webhook-url'), 'https://example2.com')
        assertAriaEnabled(getByTestId('submit-action-webhook'))

        userEvent.click(getByTestId('submit-action-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorWebhook',
            enabled: true,
            includeResults: false,
            id: '1',
            url: 'https://example2.com',
        })
    })

    test('open and delete', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <WebhookAction
                    {...props}
                    action={{
                        __typename: 'MonitorWebhook',
                        enabled: true,
                        includeResults: false,
                        id: '2',
                        url: 'https://example.com',
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-webhook'))
        userEvent.click(getByTestId('delete-action-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, undefined)
    })

    test('enable and disable', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <WebhookAction
                    {...props}
                    action={{
                        __typename: 'MonitorWebhook',
                        enabled: false,
                        includeResults: false,
                        id: '5',
                        url: 'https://example.com',
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        expect(getByTestId('enable-action-toggle-collapsed-webhook')).not.toBeChecked()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-webhook'))
        expect(getByTestId('enable-action-toggle-collapsed-webhook')).toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorWebhook',
            enabled: true,
            includeResults: false,
            id: '5',
            url: 'https://example.com',
        })

        setActionSpy.resetHistory()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-webhook'))
        expect(getByTestId('enable-action-toggle-collapsed-webhook')).not.toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorWebhook',
            enabled: false,
            includeResults: false,
            id: '5',
            url: 'https://example.com',
        })
    })

    test('open, edit, cancel, open again', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = renderWithBrandedContext(
            <MockedTestProvider>
                <WebhookAction
                    {...props}
                    action={{
                        __typename: 'MonitorWebhook',
                        enabled: true,
                        includeResults: false,
                        id: '5',
                        url: 'https://example.com',
                    }}
                    setAction={setActionSpy}
                />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-webhook'))

        expect(getByTestId('enable-action-toggle-expanded-webhook')).toBeChecked()
        userEvent.click(getByTestId('enable-action-toggle-expanded-webhook'))
        expect(getByTestId('enable-action-toggle-expanded-webhook')).not.toBeChecked()

        userEvent.type(getByTestId('webhook-url'), 'https://example2.com')

        userEvent.click(getByTestId('cancel-action-webhook'))

        userEvent.click(getByTestId('form-action-toggle-webhook'))
        expect(getByTestId('webhook-url')).toHaveValue('https://example.com')
        expect(getByTestId('enable-action-toggle-expanded-webhook')).toBeChecked()

        sinon.assert.notCalled(setActionSpy)
    })

    describe('Send test message', () => {
        const mockAction: MonitorAction = {
            __typename: 'MonitorWebhook',
            enabled: false,
            includeResults: false,
            id: '5',
            url: 'https://example.com',
        }

        const mockedVariables: SendTestWebhookVariables = {
            namespace: props.authenticatedUser.id,
            description: props.monitorName,
            webhook: {
                enabled: true,
                includeResults: false,
                url: mockAction.url,
            },
        }

        test('disabled if no webhook url set', () => {
            const { getByTestId } = renderWithBrandedContext(
                <MockedTestProvider>
                    <WebhookAction {...props} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-webhook'))
            assertAriaDisabled(getByTestId('send-test-webhook'))
        })

        test('disabled if no monitor name set', () => {
            const { getByTestId } = renderWithBrandedContext(
                <MockedTestProvider>
                    <WebhookAction {...props} monitorName="" />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-webhook'))
            assertAriaDisabled(getByTestId('send-test-webhook'))
        })

        test('send test message, success', async () => {
            const mockedResponse: MockedResponse<SendTestWebhookResult> = {
                request: {
                    query: SEND_TEST_WEBHOOK,
                    variables: mockedVariables,
                },
                result: { data: { triggerTestWebhookAction: { alwaysNil: null } } },
            }

            const { getByTestId, queryByTestId } = renderWithBrandedContext(
                <MockedTestProvider mocks={[mockedResponse]}>
                    <WebhookAction {...props} action={mockAction} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-webhook'))
            expect(getByTestId('send-test-webhook')).toHaveTextContent('Call webhook with test payload')

            userEvent.click(getByTestId('send-test-webhook'))
            expect(getByTestId('send-test-webhook')).toHaveTextContent('Calling webhook...')

            await waitForNextApolloResponse()

            expect(getByTestId('send-test-webhook')).toHaveTextContent('Test call completed!')
            assertAriaDisabled(getByTestId('send-test-webhook'))

            expect(queryByTestId('send-test-webhook')).toBeInTheDocument()
            expect(queryByTestId('test-email-webhook')).not.toBeInTheDocument()
        })

        test('send test message, error', async () => {
            const mockedResponse: MockedResponse<SendTestWebhookResult> = {
                request: {
                    query: SEND_TEST_WEBHOOK,
                    variables: mockedVariables,
                },
                error: new Error('An error occurred'),
            }

            const { getByTestId, queryByTestId } = renderWithBrandedContext(
                <MockedTestProvider mocks={[mockedResponse]}>
                    <WebhookAction {...props} action={mockAction} />
                </MockedTestProvider>
            )

            userEvent.click(getByTestId('form-action-toggle-webhook'))
            expect(getByTestId('send-test-webhook')).toHaveTextContent('Call webhook with test payload')

            userEvent.click(getByTestId('send-test-webhook'))

            await waitForNextApolloResponse()

            expect(getByTestId('send-test-webhook')).toHaveTextContent('Call webhook with test payload')

            assertAriaEnabled(getByTestId('send-test-webhook'))

            expect(queryByTestId('send-test-webhook-again')).not.toBeInTheDocument()
            expect(queryByTestId('test-webhook-error')).toBeInTheDocument()
        })
    })
})
