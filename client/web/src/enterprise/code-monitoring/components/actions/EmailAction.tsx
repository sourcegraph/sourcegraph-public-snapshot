import { gql, useMutation } from '@apollo/client'
import classNames from 'classnames'
import { noop } from 'lodash'
import React, { useState, useCallback } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { MonitorEmailPriority, SendTestEmailResult, SendTestEmailVariables } from '../../../../graphql-operations'
import { ActionProps } from '../FormActionArea'
import styles from '../FormActionArea.module.scss'

import { ActionEditor } from './ActionEditor'

export const SEND_TEST_EMAIL = gql`
    mutation SendTestEmail($namespace: ID!, $description: String!, $email: MonitorEmailInput!) {
        triggerTestEmailAction(namespace: $namespace, description: $description, email: $email) {
            alwaysNil
        }
    }
`

export const EmailAction: React.FunctionComponent<ActionProps> = ({
    action,
    setAction,
    disabled,
    authenticatedUser,
    monitorName,
    _testStartOpen,
}) => {
    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(action ? action.enabled : true)

    const toggleEmailNotificationEnabled: (enabled: boolean) => void = useCallback(
        enabled => {
            setEmailNotificationEnabled(enabled)
            setAction({
                __typename: 'MonitorEmail',
                id: action?.id ?? '',
                recipients: { nodes: [{ id: authenticatedUser.id }] },
                enabled,
                includeResults: false,
            })
        },
        [action?.id, authenticatedUser.id, setAction]
    )

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            if (!action) {
                // We are creating a new monitor if there are no actions yet.
                // The ID can be empty here, since we'll generate a new ID when we send the creation request.
                setAction({
                    __typename: 'MonitorEmail',
                    id: '',
                    recipients: { nodes: [{ id: authenticatedUser.id }] },
                    enabled: true,
                    includeResults: false,
                })
            }
        },
        [action, authenticatedUser.id, setAction]
    )

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    const [sendTestEmail, { loading, error, called }] = useMutation<SendTestEmailResult, SendTestEmailVariables>(
        SEND_TEST_EMAIL
    )
    const isSendTestEmailButtonDisabled = loading || (called && !error) || !monitorName

    const onSendTestEmail = useCallback(() => {
        sendTestEmail({
            variables: {
                namespace: authenticatedUser.id,
                description: monitorName,
                email: {
                    enabled: true,
                    includeResults: false,
                    priority: MonitorEmailPriority.NORMAL,
                    recipients: [authenticatedUser.id],
                    header: '',
                },
            },
        }).catch(noop) // Ignore errors, they will be handled with the error state from useMutation
    }, [authenticatedUser.id, monitorName, sendTestEmail])

    const sendTestEmailButtonText = loading
        ? 'Sending email...'
        : called && !error
        ? 'Test email sent!'
        : 'Send test email'

    return (
        <ActionEditor
            title="Send email notifications"
            label="Send email notifications"
            subtitle="Deliver email notifications to specified recipients."
            idName="email"
            disabled={disabled}
            completed={!!action}
            completedSubtitle={authenticatedUser.email}
            actionEnabled={emailNotificationEnabled}
            toggleActionEnabled={toggleEmailNotificationEnabled}
            onSubmit={onSubmit}
            canDelete={!!action}
            onDelete={onDelete}
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
            <div className="flex mt-1">
                <Button
                    className="mr-2"
                    variant="secondary"
                    outline={!isSendTestEmailButtonDisabled}
                    disabled={isSendTestEmailButtonDisabled}
                    onClick={onSendTestEmail}
                    size="sm"
                    data-testid="send-test-email"
                >
                    {sendTestEmailButtonText}
                </Button>
                {called && !error && !loading && monitorName && (
                    <Button
                        className="p-0"
                        onClick={onSendTestEmail}
                        variant="link"
                        size="sm"
                        data-testid="send-test-email-again"
                    >
                        Send again
                    </Button>
                )}
                {!monitorName && (
                    <small className={classNames('mt-2 form-text', styles.testActionError)}>
                        Please provide a name for the code monitor before sending a test
                    </small>
                )}
                {error && (
                    <small
                        className={classNames('mt-2 form-text', styles.testActionError)}
                        data-testid="test-email-error"
                    >
                        {error.message}
                    </small>
                )}
            </div>
        </ActionEditor>
    )
}
