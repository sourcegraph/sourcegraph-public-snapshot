import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import sinon from 'sinon'

import { ActionProps } from '../FormActionArea'

import { SlackWebhookAction } from './SlackWebhookAction'

describe('SlackWebhookAction', () => {
    const props: ActionProps = {
        action: undefined,
        setAction: sinon.stub(),
        disabled: false,
        monitorName: 'Test',
    }

    test('open and submit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(<SlackWebhookAction {...props} setAction={setActionSpy} />)

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))

        expect(getByTestId('submit-action-slack-webhook')).toBeDisabled()

        userEvent.type(getByTestId('slack-webhook-url'), 'https://example.com')
        expect(getByTestId('submit-action-slack-webhook')).toBeEnabled()

        userEvent.click(getByTestId('submit-action-slack-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: true,
            includeResults: false,
            id: '',
            url: 'https://example.com',
        })
    })

    test('open and edit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <SlackWebhookAction
                {...props}
                setAction={setActionSpy}
                action={{
                    __typename: 'MonitorSlackWebhook',
                    enabled: true,
                    includeResults: false,
                    id: '1',
                    url: 'https://example.com',
                }}
            />
        )

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
        expect(getByTestId('submit-action-slack-webhook')).toBeEnabled()

        userEvent.clear(getByTestId('slack-webhook-url'))
        expect(getByTestId('submit-action-slack-webhook')).toBeDisabled()

        userEvent.type(getByTestId('slack-webhook-url'), 'https://example2.com')
        expect(getByTestId('submit-action-slack-webhook')).toBeEnabled()

        userEvent.click(getByTestId('submit-action-slack-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: true,
            includeResults: false,
            id: '1',
            url: 'https://example2.com',
        })
    })

    test('open and delete', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <SlackWebhookAction
                {...props}
                action={{
                    __typename: 'MonitorSlackWebhook',
                    enabled: true,
                    includeResults: false,
                    id: '2',
                    url: 'https://example.com',
                }}
                setAction={setActionSpy}
            />
        )

        userEvent.click(getByTestId('form-action-toggle-slack-webhook'))
        userEvent.click(getByTestId('delete-action-slack-webhook'))

        sinon.assert.calledOnceWithExactly(setActionSpy, undefined)
    })

    test('enable and disable', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
            <SlackWebhookAction
                {...props}
                action={{
                    __typename: 'MonitorSlackWebhook',
                    enabled: false,
                    includeResults: false,
                    id: '5',
                    url: 'https://example.com',
                }}
                setAction={setActionSpy}
            />
        )

        expect(getByTestId('enable-action-toggle-collapsed-slack-webhook')).not.toBeChecked()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-slack-webhook'))
        expect(getByTestId('enable-action-toggle-collapsed-slack-webhook')).toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: true,
            includeResults: false,
            id: '5',
            url: 'https://example.com',
        })

        setActionSpy.resetHistory()

        userEvent.click(getByTestId('enable-action-toggle-collapsed-slack-webhook'))
        expect(getByTestId('enable-action-toggle-collapsed-slack-webhook')).not.toBeChecked()
        sinon.assert.calledOnceWithExactly(setActionSpy, {
            __typename: 'MonitorSlackWebhook',
            enabled: false,
            includeResults: false,
            id: '5',
            url: 'https://example.com',
        })
    })
})
