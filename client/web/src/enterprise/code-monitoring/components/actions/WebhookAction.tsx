import React, { useCallback, useMemo, useState } from 'react'

import { gql, useMutation } from '@apollo/client'
import { noop } from 'lodash'

import { Alert, Input, Link, ProductStatusBadge, Label } from '@sourcegraph/wildcard'

import type { SendTestWebhookResult, SendTestWebhookVariables } from '../../../../graphql-operations'
import type { ActionProps } from '../FormActionArea'

import { ActionEditor } from './ActionEditor'

export const SEND_TEST_WEBHOOK = gql`
    mutation SendTestWebhook($namespace: ID!, $description: String!, $webhook: MonitorWebhookInput!) {
        triggerTestWebhookAction(namespace: $namespace, description: $description, webhook: $webhook) {
            alwaysNil
        }
    }
`

export const WebhookAction: React.FunctionComponent<React.PropsWithChildren<ActionProps>> = ({
    action,
    setAction,
    disabled,
    monitorName,
    authenticatedUser,
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

    const [url, setUrl] = useState(action && action.__typename === 'MonitorWebhook' ? action.url : '')
    const urlIsValid = useMemo(() => !!url.match(/^https?:\/\//), [url])

    const [includeResults, setIncludeResults] = useState(action ? action.includeResults : false)
    const toggleIncludeResults: (includeResults: boolean) => void = useCallback(includeResults => {
        setIncludeResults(includeResults)
    }, [])

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorWebhook',
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
        setUrl(action && action.__typename === 'MonitorWebhook' ? action.url : '')
        setIncludeResults(action ? action.includeResults : false)
    }, [action])

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    const [sendTestMessage, { loading, error, called }] = useMutation<SendTestWebhookResult, SendTestWebhookVariables>(
        SEND_TEST_WEBHOOK
    )

    const onSendTestMessage = useCallback(() => {
        sendTestMessage({
            variables: {
                namespace: authenticatedUser.id,
                description: monitorName,
                webhook: { url, enabled: true, includeResults },
            },
        }).catch(noop) // Ignore errors, they will be handled with the error state from useMutation
    }, [authenticatedUser.id, includeResults, monitorName, sendTestMessage, url])

    const testButtonText = loading
        ? 'Calling webhook...'
        : called && !error
        ? 'Test call completed!'
        : 'Call webhook with test payload'

    const testButtonDisabledReason = !monitorName
        ? 'Please provide a name for the code monitor before making a test call'
        : !url
        ? 'Please provide a webhook URL before making a test call'
        : undefined

    const testState = loading ? 'loading' : called && !error ? 'called' : error || undefined

    return (
        <ActionEditor
            title={
                <div>
                    Call a webhook <ProductStatusBadge className="ml-1 mb-1" status="beta" />{' '}
                </div>
            }
            subtitle="Calls the specified URL with a JSON payload."
            idName="webhook"
            disabled={disabled}
            completed={!!action}
            completedSubtitle="The webhook at the specified URL will be called."
            actionEnabled={enabled}
            toggleActionEnabled={toggleWebhookEnabled}
            canSubmit={!!urlIsValid}
            includeResults={includeResults}
            toggleIncludeResults={toggleIncludeResults}
            onSubmit={onSubmit}
            onCancel={onCancel}
            canDelete={!!action}
            onDelete={onDelete}
            testState={testState}
            testButtonDisabledReason={testButtonDisabledReason}
            testButtonText={testButtonText}
            testAgainButtonText="Test again"
            onTest={onSendTestMessage}
            _testStartOpen={_testStartOpen}
        >
            <Alert aria-live="off" variant="info" className="mt-4">
                The specified webhook URL will be called with a JSON payload.
                <br />
                <Link to="/help/code_monitoring/how-tos/webhook" target="_blank" rel="noopener">
                    Read more about how to set up webhooks and the JSON schema in the docs.
                </Link>
            </Alert>
            <div className="form-group">
                <Label htmlFor="code-monitor-webhook-url">Webhook URL</Label>
                <Input
                    id="code-monitor-webhook-url"
                    type="url"
                    className="mb-2"
                    data-testid="webhook-url"
                    required={true}
                    onChange={event => {
                        setUrl(event.target.value)
                    }}
                    value={url}
                    autoFocus={true}
                    spellCheck={false}
                    status={urlIsValid ? 'valid' : url ? 'error' : undefined /* Don't show error state when empty */}
                    error={!urlIsValid && url ? 'Enter a valid webhook URL.' : undefined}
                />
            </div>
        </ActionEditor>
    )
}
