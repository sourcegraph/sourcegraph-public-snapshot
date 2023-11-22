import React, { useCallback, useMemo, useState } from 'react'

import { gql, useMutation } from '@apollo/client'
import { noop } from 'lodash'

import { Alert, Input, Link, ProductStatusBadge, Label } from '@sourcegraph/wildcard'

import type { SendTestSlackWebhookResult, SendTestSlackWebhookVariables } from '../../../../graphql-operations'
import type { ActionProps } from '../FormActionArea'

import { ActionEditor } from './ActionEditor'

export const SEND_TEST_SLACK_WEBHOOK = gql`
    mutation SendTestSlackWebhook($namespace: ID!, $description: String!, $slackWebhook: MonitorSlackWebhookInput!) {
        triggerTestSlackWebhookAction(namespace: $namespace, description: $description, slackWebhook: $slackWebhook) {
            alwaysNil
        }
    }
`

export const SlackWebhookAction: React.FunctionComponent<React.PropsWithChildren<ActionProps>> = ({
    action,
    setAction,
    disabled,
    authenticatedUser,
    monitorName,
    _testStartOpen,
}) => {
    const [enabled, setEnabled] = useState(action ? action.enabled : true)

    const toggleWebhookEnabled: (enabled: boolean, saveImmediately: boolean) => void = useCallback(
        (enabled, saveImmediately) => {
            setEnabled(enabled)
            if (action && saveImmediately) {
                setAction({ ...action, enabled })
            }
        },
        [action, setAction]
    )

    const [url, setUrl] = useState(action && action.__typename === 'MonitorSlackWebhook' ? action.url : '')
    const urlIsValid = useMemo(() => url.startsWith('https://hooks.slack.com/services/'), [url])

    const [includeResults, setIncludeResults] = useState(action ? action.includeResults : false)
    const toggleIncludeResults: (includeResults: boolean) => void = useCallback(includeResults => {
        setIncludeResults(includeResults)
    }, [])

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorSlackWebhook',
                id: action ? action.id : '',
                url,
                enabled,
                includeResults,
            })
        },
        [action, includeResults, setAction, url, enabled]
    )

    const onCancel: React.FormEventHandler = useCallback(() => {
        setEnabled(action ? action.enabled : true)
        setUrl(action && action.__typename === 'MonitorSlackWebhook' ? action.url : '')
        setIncludeResults(action ? action.includeResults : false)
    }, [action])

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    const [sendTestMessage, { loading, error, called }] = useMutation<
        SendTestSlackWebhookResult,
        SendTestSlackWebhookVariables
    >(SEND_TEST_SLACK_WEBHOOK)

    const onSendTestMessage = useCallback(() => {
        sendTestMessage({
            variables: {
                namespace: authenticatedUser.id,
                description: monitorName,
                slackWebhook: { url, enabled: true, includeResults },
            },
        }).catch(noop) // Ignore errors, they will be handled with the error state from useMutation
    }, [authenticatedUser.id, includeResults, monitorName, sendTestMessage, url])

    const testButtonText = loading
        ? 'Sending message...'
        : called && !error
        ? 'Test message sent!'
        : 'Send test message'

    const testButtonDisabledReason = !monitorName
        ? 'Please provide a name for the code monitor before sending a test'
        : !url
        ? 'Please provide a webhook URL before sending a test'
        : undefined

    const testState = loading ? 'loading' : called && !error ? 'called' : error || undefined

    return (
        <ActionEditor
            title={
                <div>
                    Send Slack message to channel <ProductStatusBadge className="ml-1 mb-1" status="beta" />{' '}
                </div>
            }
            subtitle="Post to a specified Slack channel. Requires webhook configuration."
            idName="slack-webhook"
            disabled={disabled}
            completed={!!action}
            completedSubtitle="Notification will be sent to the specified Slack webhook URL."
            actionEnabled={enabled}
            toggleActionEnabled={toggleWebhookEnabled}
            canSubmit={urlIsValid}
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
            onTest={onSendTestMessage}
            _testStartOpen={_testStartOpen}
        >
            <Alert aria-live="off" variant="info" className="mt-4">
                Go to{' '}
                <Link to="https://api.slack.com/apps" target="_blank" rel="noopener">
                    Slack
                </Link>{' '}
                to create a webhook URL.
                <br />
                <Link to="/help/code_monitoring/how-tos/slack" target="_blank" rel="noopener">
                    Read more about how to set up Slack webhooks in the docs.
                </Link>
            </Alert>
            <div className="form-group">
                <Label htmlFor="code-monitor-slack-webhook-url">Slack webhook URL</Label>
                <Input
                    id="code-monitor-slack-webhook-url"
                    type="url"
                    className="mb-2"
                    data-testid="slack-webhook-url"
                    required={true}
                    onChange={event => {
                        setUrl(event.target.value)
                    }}
                    value={url}
                    autoFocus={true}
                    spellCheck={false}
                    status={urlIsValid ? 'valid' : url ? 'error' : undefined /* Don't show error state when empty */}
                    error={!urlIsValid && url ? 'Enter a valid Slack webhook URL.' : undefined}
                />
            </div>
        </ActionEditor>
    )
}
