import { render } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import { NEVER, of, throwError } from 'rxjs'
import sinon from 'sinon'

import { mockAuthenticatedUser } from '../../testing/util'

import { EmailAction, EmailActionProps } from './EmailAction'

describe('EmailAction', () => {
    const props: EmailActionProps = {
        action: undefined,
        setAction: sinon.stub(),
        disabled: false,
        authenticatedUser: mockAuthenticatedUser,
        monitorName: 'Test',
        triggerTestEmailAction: sinon.stub(),
    }

    test('open and submit', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(<EmailAction {...props} setAction={setActionSpy} />)

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

    test('open and delete', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
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
        )

        userEvent.click(getByTestId('form-action-toggle-email'))
        userEvent.click(getByTestId('delete-action-email'))

        sinon.assert.calledOnceWithExactly(setActionSpy, undefined)
    })

    test('enable and disable', () => {
        const setActionSpy = sinon.spy()
        const { getByTestId } = render(
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

    describe('Send test email', () => {
        let clock: sinon.SinonFakeTimers
        beforeEach(() => {
            clock = sinon.useFakeTimers()
        })
        afterEach(() => {
            clock.restore()
        })

        test('disabled if no monitor name set', () => {
            const { getByTestId } = render(<EmailAction {...props} monitorName="" />)

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toBeDisabled()
        })

        test('send test email, loading', () => {
            const { getByTestId, queryByTestId } = render(
                <EmailAction {...props} triggerTestEmailAction={() => NEVER} />
            )

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            userEvent.click(getByTestId('send-test-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Sending email...')

            clock.tick(1000)

            expect(getByTestId('send-test-email')).toHaveTextContent('Sending email...')
            expect(getByTestId('send-test-email')).toBeDisabled()

            expect(queryByTestId('send-test-email-again')).not.toBeInTheDocument()
            expect(queryByTestId('test-email-error')).not.toBeInTheDocument()
        })

        test('send test email, success', () => {
            const { getByTestId, queryByTestId } = render(
                <EmailAction {...props} triggerTestEmailAction={() => of(undefined)} />
            )

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            userEvent.click(getByTestId('send-test-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Sending email...')

            clock.tick(1000)

            expect(getByTestId('send-test-email')).toHaveTextContent('Test email sent!')
            expect(getByTestId('send-test-email')).toBeDisabled()

            expect(queryByTestId('send-test-email-again')).toBeInTheDocument()
            expect(queryByTestId('test-email-error')).not.toBeInTheDocument()
        })

        test('send test email, error', () => {
            const { getByTestId, queryByTestId } = render(
                <EmailAction {...props} triggerTestEmailAction={() => throwError(new Error('error'))} />
            )

            userEvent.click(getByTestId('form-action-toggle-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            userEvent.click(getByTestId('send-test-email'))
            expect(getByTestId('send-test-email')).toHaveTextContent('Send test email')

            clock.tick(1000)

            expect(getByTestId('send-test-email')).toBeEnabled()

            expect(queryByTestId('send-test-email-again')).not.toBeInTheDocument()
            expect(queryByTestId('test-email-error')).toBeInTheDocument()
        })
    })
})
