import classnames from 'classnames'
import React, { useCallback, useState } from 'react'
import { Observable } from 'rxjs'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'
import { AuthenticatedUser } from '../../../auth'
import { MonitorEmailPriority } from '../../../graphql-operations'
import * as H from 'history'
import { Toggle } from '../../../../../branded/src/components/Toggle'
import { useEventObservable } from '../../../../../shared/src/util/useObservable'
import { createCodeMonitor } from '../backend'
import { FormActionArea } from './FormActionArea'
import { FormTriggerArea } from './FormTriggerArea'
import { mergeMap, startWith, catchError, tap } from 'rxjs/operators'
import { Form } from '../../../../../branded/src/components/Form'

export interface CodeMonitorFormProps {
    location: H.Location
    authenticatedUser: AuthenticatedUser
}

interface FormCompletionSteps {
    triggerCompleted: boolean
    actionCompleted: boolean
}

export interface Action {
    recipient: string
    enabled: boolean
}

export interface CodeMonitorFields {
    description: string
    query: string
    enabled: boolean
    action: Action
}

export const CodeMonitorForm: React.FunctionComponent<CodeMonitorFormProps> = props => {
    const LOADING = 'loading' as const

    const [codeMonitor, setCodeMonitor] = useState<CodeMonitorFields>({
        description: '',
        query: '',
        action: { recipient: props.authenticatedUser.id, enabled: true },
        enabled: true,
    })
    const onNameChange = useCallback(
        (description: string): void => setCodeMonitor(codeMonitor => ({ ...codeMonitor, description })),
        []
    )
    const onQueryChange = useCallback(
        (query: string): void => setCodeMonitor(codeMonitor => ({ ...codeMonitor, query })),
        []
    )
    const onEnabledChange = useCallback(
        (enabled: boolean): void => setCodeMonitor(codeMonitor => ({ ...codeMonitor, enabled })),
        []
    )
    const onActionChange = useCallback(
        (action: Action): void => setCodeMonitor(codeMonitor => ({ ...codeMonitor, action })),
        []
    )

    const [formCompletion, setFormCompletion] = useState<FormCompletionSteps>({
        triggerCompleted: false,
        actionCompleted: false,
    })
    const setTriggerCompleted = useCallback(() => {
        setFormCompletion(previousState => ({ ...previousState, triggerCompleted: !previousState.triggerCompleted }))
    }, [])
    const setActionCompleted = useCallback(() => {
        setFormCompletion(previousState => ({ ...previousState, actionCompleted: !previousState.actionCompleted }))
    }, [])

    const [createRequest, codeMonitorOrError] = useEventObservable(
        useCallback(
            (submit: Observable<React.FormEvent<HTMLFormElement>>) =>
                submit.pipe(
                    tap(event => event.preventDefault()),
                    mergeMap(() =>
                        createCodeMonitor({
                            monitor: {
                                namespace: props.authenticatedUser.id,
                                description: codeMonitor.description,
                                enabled: codeMonitor.enabled,
                            },
                            trigger: { query: codeMonitor.query },
                            actions: [
                                {
                                    email: {
                                        enabled: codeMonitor.action.enabled,
                                        priority: MonitorEmailPriority.NORMAL,
                                        recipients: [props.authenticatedUser.id],
                                        header: '',
                                    },
                                },
                            ],
                        }).pipe(
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [props.authenticatedUser, codeMonitor]
        )
    )

    return (
        <Form className="my-4" onSubmit={createRequest}>
            <div className="flex mb-4">
                Name
                <div>
                    <input
                        type="text"
                        className="form-control my-2 test-name-input"
                        required={true}
                        onChange={event => {
                            onNameChange(event.target.value)
                        }}
                        autoFocus={true}
                    />
                </div>
                <small className="text-muted">
                    Give it a short, descriptive name to reference events on Sourcegraph and in notifications. Do not
                    include:{' '}
                    <a href="" target="_blank" rel="noopener">
                        {/* TODO: populate link */}
                        confidential information
                    </a>
                    .
                </small>
            </div>
            <div className="flex">
                Owner
                <select className="form-control my-2 w-auto" disabled={true}>
                    <option value={props.authenticatedUser.displayName || props.authenticatedUser.username}>
                        {props.authenticatedUser.username}
                    </option>
                </select>
                <small className="text-muted">Event history and configuration will not be shared.</small>
            </div>
            <hr className="my-4" />
            <div className="create-monitor-page__triggers mb-4">
                <FormTriggerArea
                    query={codeMonitor.query}
                    onQueryChange={onQueryChange}
                    triggerCompleted={formCompletion.triggerCompleted}
                    setTriggerCompleted={setTriggerCompleted}
                />
            </div>
            <div
                className={classnames({
                    'create-monitor-page__actions--disabled': !formCompletion.triggerCompleted,
                })}
            >
                <FormActionArea
                    setActionCompleted={setActionCompleted}
                    actionCompleted={formCompletion.actionCompleted}
                    authenticatedUser={props.authenticatedUser}
                    disabled={!formCompletion.triggerCompleted}
                    onActionChange={onActionChange}
                />
            </div>
            <div>
                <div className="d-flex my-4">
                    <div>
                        <Toggle
                            title="Active"
                            value={codeMonitor.enabled}
                            onToggle={onEnabledChange}
                            className="mr-2"
                        />{' '}
                    </div>
                    <div className="flex-column">
                        <div>Active</div>
                        <div className="text-muted">We will watch for the trigger and run actions in response</div>
                    </div>
                </div>
                <div className="flex my-4">
                    <button
                        type="submit"
                        disabled={
                            !formCompletion.actionCompleted ||
                            isErrorLike(codeMonitorOrError) ||
                            codeMonitorOrError === LOADING
                        }
                        className="btn btn-primary mr-2 test-submit-monitor"
                    >
                        Create code monitor
                    </button>
                    <button type="button" className="btn btn-outline-secondary">
                        {/* TODO: this should link somewhere */}
                        Cancel
                    </button>
                </div>
                {/** TODO: Error and success states. We will probably redirect the user to another page, so we could remove the success state. */}
                {!isErrorLike(codeMonitorOrError) && !!codeMonitorOrError && codeMonitorOrError !== LOADING && (
                    <div className="alert alert-success">
                        Successfully created monitor {codeMonitorOrError.description}
                    </div>
                )}
                {isErrorLike(codeMonitorOrError) && (
                    <div className="alert alert-danger">Failed to create monitor: {codeMonitorOrError.message}</div>
                )}
            </div>
        </Form>
    )
}
