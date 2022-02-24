import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import sinon from 'sinon'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { mockAuthenticatedUser } from '../../testing/util'
import { ActionProps } from '../FormActionArea'

import { WebhookAction } from './WebhookAction'

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
        const { getByTestId } = render(
            <MockedTestProvider>
                <WebhookAction {...props} setAction={setActionSpy} />
            </MockedTestProvider>
        )

        userEvent.click(getByTestId('form-action-toggle-webhook'))

        expect(getByTestId('submit-action-webhook')).toBeDisabled()

        userEvent.type(getByTestId('webhook-url'), 'https://example.com')
        expect(getByTestId('submit-action-webhook')).toBeEnabled()

        userEvent.click(getByTestId('submit-action-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorWebhook',
            enabled: true,
            includeResults: false,
            id: '',
            url: 'https://example.com',
        })
    })

    test('open and edit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
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
        expect(getByTestId('submit-action-webhook')).toBeEnabled()

        userEvent.clear(getByTestId('webhook-url'))
        expect(getByTestId('submit-action-webhook')).toBeDisabled()

        userEvent.type(getByTestId('webhook-url'), 'https://example2.com')
        expect(getByTestId('submit-action-webhook')).toBeEnabled()

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
        const { getByTestId } = render(
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
        const { getByTestId } = render(
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
})
