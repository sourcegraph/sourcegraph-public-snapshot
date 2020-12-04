import classnames from 'classnames'
import React, { useCallback, useState } from 'react'
import { Observable } from 'rxjs'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'
import { AuthenticatedUser } from '../../../auth'
import * as H from 'history'
import { Toggle } from '../../../../../branded/src/components/Toggle'
import { FormActionArea } from './FormActionArea'
import { FormTriggerArea } from './FormTriggerArea'
import { mergeMap, startWith, catchError, tap } from 'rxjs/operators'
import { Form } from '../../../../../branded/src/components/Form'
import { useEventObservable } from '../../../../../shared/src/util/useObservable'
import { CodeMonitorFields } from '../../../graphql-operations'

export interface CodeMonitorFormProps {
    location: H.Location
    authenticatedUser: AuthenticatedUser
    /**
     * A function that takes in a code monitor and emits an Observable with all or some
     * of the CodeMonitorFields when the form is submitted.
     */
    onSubmit: (codeMonitor: CodeMonitorFields) => Observable<Partial<CodeMonitorFields>>
    /** The text for the submit button. */
    submitButtonLabel: string
}

interface FormCompletionSteps {
    triggerCompleted: boolean
    actionCompleted: boolean
}

export const CodeMonitorForm: React.FunctionComponent<CodeMonitorFormProps> = ({
    authenticatedUser,
    onSubmit,
    submitButtonLabel,
}) => {
    const LOADING = 'loading' as const

    const [currentCodeMonitorState, setCodeMonitor] = useState<CodeMonitorFields>({
        id: '',
        description: '',
        enabled: true,
        trigger: { id: '', query: '' },
        actions: {
            nodes: [],
        },
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
    const onActionsChange = useCallback(
        (actions: CodeMonitorFields['actions']): void => setCodeMonitor(codeMonitor => ({ ...codeMonitor, actions })),
        []
    )

    const [formCompletion, setFormCompletion] = useState<FormCompletionSteps>({
        triggerCompleted: false,
        actionCompleted: false,
    })
    const setTriggerCompleted = useCallback((complete: boolean) => {
        setFormCompletion(previousState => ({ ...previousState, triggerCompleted: complete }))
    }, [])
    const setActionsCompleted = useCallback((complete: boolean) => {
        setFormCompletion(previousState => ({ ...previousState, actionCompleted: complete }))
    }, [])

    const [requestOnSubmit, codeMonitorOrError] = useEventObservable(
        useCallback(
            (submit: Observable<React.FormEvent<HTMLFormElement>>) =>
                submit.pipe(
                    tap(event => event.preventDefault()),
                    mergeMap(() =>
                        onSubmit(currentCodeMonitorState).pipe(
                            startWith(LOADING),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [onSubmit, currentCodeMonitorState]
        )
    )

    return (
        <Form className="my-4" onSubmit={requestOnSubmit}>
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
                    <option value={authenticatedUser.displayName || authenticatedUser.username}>
                        {authenticatedUser.username}
                    </option>
                </select>
                <small className="text-muted">Event history and configuration will not be shared.</small>
            </div>
            <hr className="my-4" />
            <div className="create-monitor-page__triggers mb-4">
                <FormTriggerArea
                    query={currentCodeMonitorState.trigger.query}
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
                    actions={currentCodeMonitorState.actions}
                    setActionsCompleted={setActionsCompleted}
                    actionsCompleted={formCompletion.actionCompleted}
                    authenticatedUser={authenticatedUser}
                    disabled={!formCompletion.triggerCompleted}
                    onActionsChange={onActionsChange}
                />
            </div>
            <div>
                <div className="d-flex my-4">
                    <div>
                        <Toggle
                            title="Active"
                            value={currentCodeMonitorState.enabled}
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
                        {submitButtonLabel}
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
