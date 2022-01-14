import React, { useEffect, useState } from 'react'

import { AuthenticatedUser } from '../../../auth'
import { CodeMonitorFields } from '../../../graphql-operations'

import { EmailAction } from './actions/EmailAction'

export interface ActionAreaProps {
    actions: CodeMonitorFields['actions']
    actionsCompleted: boolean
    setActionsCompleted: (completed: boolean) => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    onActionsChange: (action: CodeMonitorFields['actions']) => void
    description: string
    cardClassName?: string
    cardBtnClassName?: string
    cardLinkClassName?: string
}

export type MonitorAction = CodeMonitorFields['actions']['nodes'][number]

/**
 * TODO farhan: this component is built with the assumption that each monitor has exactly one email action.
 * Refactor to accomodate for more than one.
 */
export const FormActionArea: React.FunctionComponent<ActionAreaProps> = ({
    actions,
    actionsCompleted,
    setActionsCompleted,
    disabled,
    authenticatedUser,
    onActionsChange,
    description,
    cardClassName,
    cardBtnClassName,
    cardLinkClassName,
}) => {
    const [emailAction, setEmailAction] = useState<MonitorAction | undefined>(
        actions.nodes.find(action => action.__typename === 'MonitorEmail')
    )
    const [emailActionCompleted, setEmailActionCompleted] = useState(!!emailAction && actionsCompleted)

    // Form is completed only if all set actions are completed and all incomplete actions are unset,
    // and there is at least one completed action.
    // (Currently there is only one action, but more will be added.)
    useEffect(() => {
        setActionsCompleted(!!emailAction && emailActionCompleted)
    }, [emailAction, emailActionCompleted, setActionsCompleted])

    useEffect(() => {
        const actions: CodeMonitorFields['actions'] = { nodes: [] }
        if (emailAction) {
            actions.nodes.push(emailAction)
        }
        onActionsChange(actions)
    }, [emailAction, onActionsChange])

    return (
        <>
            <h3 className="mb-1">Actions</h3>
            <span className="text-muted">Run any number of actions in response to an event</span>
            <EmailAction
                disabled={disabled}
                action={emailAction}
                setAction={setEmailAction}
                actionCompleted={emailActionCompleted}
                setActionCompleted={setEmailActionCompleted}
                authenticatedUser={authenticatedUser}
                cardClassName={cardClassName}
                cardBtnClassName={cardBtnClassName}
                cardLinkClassName={cardLinkClassName}
                description={description}
            />
            <small className="text-muted">
                What other actions would you like to take?{' '}
                <a href="mailto:feedback@sourcegraph.com" target="_blank" rel="noopener">
                    Share feedback.
                </a>
            </small>
        </>
    )
}
