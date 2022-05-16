import React, { useState, useCallback } from 'react'

import { gql, useMutation } from '@apollo/client'
import { noop } from 'lodash'

import { MonitorEmailPriority, SendTestEmailResult, SendTestEmailVariables } from '../../../../graphql-operations'
import { ActionProps } from '../FormActionArea'

import { ActionEditor } from './ActionEditor'

export const SEND_TEST_EMAIL = gql`
    mutation SendTestEmail($namespace: ID!, $description: String!, $email: MonitorEmailInput!) {
        triggerTestEmailAction(namespace: $namespace, description: $description, email: $email) {
            alwaysNil
        }
    }
`

export const EmailAction: React.FunctionComponent<React.PropsWithChildren<ActionProps>> = ({
    action,
    setAction,
    disabled,
    authenticatedUser,
    monitorName,
    _testStartOpen,
}) => {
    const [enabled, setEnabled] = useState(action ? action.enabled : true)

    const toggleEmailNotificationEnabled: (enabled: boolean, saveImmediately: boolean) => void = useCallback(
        (enabled, saveImmediately) => {
            setEnabled(enabled)
            if (action && saveImmediately) {
                setAction({ ...action, enabled })
            }
        },
        [action, setAction]
    )

    const [includeResults, setIncludeResults] = useState(action ? action.includeResults : false)
    const toggleIncludeResults: (includeResults: boolean) => void = useCallback(includeResults => {
        setIncludeResults(includeResults)
    }, [])

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorEmail',
                id: action ? action.id : '',
                recipients: { nodes: [{ id: authenticatedUser.id }] },
                enabled,
                includeResults,
            })
        },
        [action, authenticatedUser.id, enabled, includeResults, setAction]
    )

    const onCancel: React.FormEventHandler = useCallback(() => {
        setEnabled(action ? action.enabled : true)
        setIncludeResults(action ? action.includeResults : false)
    }, [action])

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    const [sendTestEmail, { loading, error, called }] = useMutation<SendTestEmailResult, SendTestEmailVariables>(
        SEND_TEST_EMAIL
    )

    const onSendTestEmail = useCallback(() => {
        sendTestEmail({
            variables: {
                namespace: authenticatedUser.id,
                description: monitorName,
                email: {
                    enabled,
                    includeResults,
                    priority: MonitorEmailPriority.NORMAL,
                    recipients: [authenticatedUser.id],
                    header: '',
                },
            },
        }).catch(noop) // Ignore errors, they will be handled with the error state from useMutation
    }, [authenticatedUser.id, enabled, includeResults, monitorName, sendTestEmail])

    const testButtonText = loading ? 'Sending email...' : called && !error ? 'Test email sent!' : 'Send test email'

    const testButtonDisabledReason = !monitorName
        ? 'Please provide a name for the code monitor before sending a test'
        : undefined
    const testState = loading ? 'loading' : called && !error ? 'called' : error || undefined

    return (
        <ActionEditor
            title="Send email notifications"
            label="Send email notifications"
            subtitle="Deliver email notifications to specified recipients."
            idName="email"
            disabled={disabled}
            completed={!!action}
            completedSubtitle={authenticatedUser.email}
            actionEnabled={enabled}
            toggleActionEnabled={toggleEmailNotificationEnabled}
            includeResults={includeResults}
            toggleIncludeResults={toggleIncludeResults}
            onSubmit={onSubmit}
            onCancel={onCancel}
            canDelete={!!action}
            onDelete={onDelete}
            testState={testState}
            testButtonDisabledReason={testButtonDisabledReason}
            testButtonText={testButtonText}
            testAgainButtonText="Send again"
            onTest={onSendTestEmail}
            _testStartOpen={_testStartOpen}
        >
            <div className="form-group mt-4 test-action-form-email" data-testid="action-form-email">
                <label htmlFor="code-monitoring-form-actions-recipients">Recipients</label>
                <input
                    id="code-monitoring-form-actions-recipients"
                    type="text"
                    className="form-control mb-2"
                    value={`${authenticatedUser.email || ''} (you)`}
                    disabled={true}
                    autoFocus={true}
                    required={true}
                />
                <small className="text-muted">
                    Code monitors are currently limited to sending emails to your primary email address.
                </small>
            </div>
        </ActionEditor>
    )
}
