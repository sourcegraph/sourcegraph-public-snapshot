import React, { useEffect, useState } from 'react'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { Link, H3 } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../auth'
import type { CodeMonitorFields } from '../../../graphql-operations'

import { EmailAction } from './actions/EmailAction'
import { SlackWebhookAction } from './actions/SlackWebhookAction'
import { WebhookAction } from './actions/WebhookAction'

export interface ActionAreaProps {
    actions: CodeMonitorFields['actions']
    actionsCompleted: boolean
    setActionsCompleted: (completed: boolean) => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    onActionsChange: (action: CodeMonitorFields['actions']) => void
    monitorName: string
}

export interface ActionProps {
    action?: MonitorAction
    setAction: (action?: MonitorAction) => void
    disabled: boolean
    monitorName: string
    authenticatedUser: AuthenticatedUser

    // For testing purposes only
    _testStartOpen?: boolean
}

export type MonitorAction = CodeMonitorFields['actions']['nodes'][number]

/**
 * TODO farhan: this component is built with the assumption that each monitor has exactly one email action.
 * Refactor to accomodate for more than one.
 */
export const FormActionArea: React.FunctionComponent<React.PropsWithChildren<ActionAreaProps>> = ({
    actions,
    setActionsCompleted,
    disabled,
    authenticatedUser,
    onActionsChange,
    monitorName,
}) => {
    const [emailAction, setEmailAction] = useState<MonitorAction | undefined>(
        actions.nodes.find(action => action.__typename === 'MonitorEmail')
    )

    const [slackWebhookAction, setSlackWebhookAction] = useState<MonitorAction | undefined>(
        actions.nodes.find(action => action.__typename === 'MonitorSlackWebhook')
    )

    const [webhookAction, setWebhookAction] = useState<MonitorAction | undefined>(
        actions.nodes.find(action => action.__typename === 'MonitorWebhook')
    )

    // Form is completed if there is at least one action
    useEffect(() => {
        setActionsCompleted(!!emailAction || !!slackWebhookAction || !!webhookAction)
    }, [emailAction, setActionsCompleted, slackWebhookAction, webhookAction])

    useEffect(() => {
        const actions: CodeMonitorFields['actions'] = { nodes: [] }
        if (emailAction) {
            actions.nodes.push(emailAction)
        }
        if (slackWebhookAction) {
            actions.nodes.push(slackWebhookAction)
        }
        if (webhookAction) {
            actions.nodes.push(webhookAction)
        }
        onActionsChange(actions)
    }, [emailAction, onActionsChange, slackWebhookAction, webhookAction])

    const showWebhooks = useExperimentalFeatures(features => features.codeMonitoringWebHooks)

    return (
        <>
            <H3 className="mb-1">Actions</H3>
            <span className="text-muted">Run any number of actions in response to an event</span>

            <EmailAction
                disabled={disabled}
                action={emailAction}
                setAction={setEmailAction}
                authenticatedUser={authenticatedUser}
                monitorName={monitorName}
            />

            {(showWebhooks || slackWebhookAction) && (
                <SlackWebhookAction
                    disabled={disabled}
                    action={slackWebhookAction}
                    setAction={setSlackWebhookAction}
                    monitorName={monitorName}
                    authenticatedUser={authenticatedUser}
                />
            )}

            {(showWebhooks || webhookAction) && (
                <WebhookAction
                    disabled={disabled}
                    action={webhookAction}
                    setAction={setWebhookAction}
                    monitorName={monitorName}
                    authenticatedUser={authenticatedUser}
                />
            )}

            <small className="text-muted">
                What other actions would you like to take?{' '}
                <Link to="mailto:feedback@sourcegraph.com" target="_blank" rel="noopener">
                    Share feedback.
                </Link>
            </small>
        </>
    )
}
