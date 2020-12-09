import classnames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'
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
import { isEqual } from 'lodash'

export interface CodeMonitorFormProps {
    history: H.History
    location: H.Location
    authenticatedUser: AuthenticatedUser
    /**
     * A function that takes in a code monitor and emits an Observable with all or some
     * of the CodeMonitorFields when the form is submitted.
     */
    onSubmit: (codeMonitor: CodeMonitorFields) => Observable<Partial<CodeMonitorFields>>
    /** The text for the submit button. */
    submitButtonLabel: string
    /** A code monitor to initialize the form with. */
    codeMonitor?: CodeMonitorFields
}

interface FormCompletionSteps {
    triggerCompleted: boolean
    actionCompleted: boolean
}

export const CodeMonitorForm: React.FunctionComponent<CodeMonitorFormProps> = ({
    authenticatedUser,
    onSubmit,
    history,
    submitButtonLabel,
    codeMonitor,
}) => {
    const LOADING = 'loading' as const

    const [currentCodeMonitorState, setCodeMonitor] = useState<CodeMonitorFields>(
        codeMonitor ?? {
            id: '',
            description: '',
            enabled: true,
            trigger: { id: '', query: '' },
            actions: {
                nodes: [],
            },
        }
    )

    const [formCompletion, setFormCompletion] = useState<FormCompletionSteps>({
        triggerCompleted: currentCodeMonitorState.trigger.query.length > 0,
        actionCompleted: currentCodeMonitorState.actions.nodes.length > 0,
    })
    const setTriggerCompleted = useCallback((complete: boolean) => {
        setFormCompletion(previousState => ({ ...previousState, triggerCompleted: complete }))
    }, [])
    const setActionsCompleted = useCallback((complete: boolean) => {
        setFormCompletion(previousState => ({ ...previousState, actionCompleted: complete }))
    }, [])

    const onNameChange = useCallback(
        (description: string): void => setCodeMonitor(codeMonitor => ({ ...codeMonitor, description })),
        []
    )
    const onQueryChange = useCallback(
        (query: string): void =>
            setCodeMonitor(codeMonitor => ({ ...codeMonitor, trigger: { ...codeMonitor.trigger, query } })),
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

    const [requestOnSubmit, codeMonitorOrError] = useEventObservable(
        useCallback(
            (submit: Observable<React.FormEvent<HTMLFormElement>>) =>
                submit.pipe(
                    tap(event => event.preventDefault()),
                    mergeMap(() =>
                        onSubmit(currentCodeMonitorState).pipe(
                            startWith(LOADING),
                            catchError(error => [asError(error)]),
                            tap(successOrError => {
                                if (!isErrorLike(successOrError) && successOrError !== LOADING) {
                                    history.push('/code-monitoring')
                                }
                            })
                        )
                    )
                ),
            [onSubmit, currentCodeMonitorState, history]
        )
    )

    const initialCodeMonitor = useMemo(() => codeMonitor, [codeMonitor])

    // Determine whether the form has changed. If there was no intial state (i.e. we're creating a monitor), always return
    // true.
    const hasChangedFields = useMemo(
        () => (codeMonitor ? !isEqual(initialCodeMonitor, currentCodeMonitorState) : true),
        [initialCodeMonitor, codeMonitor, currentCodeMonitorState]
    )

    const onCancel = useCallback(() => {
        if (hasChangedFields) {
            if (window.confirm('Leave page? All unsaved changes will be lost.')) {
                history.push('/code-monitoring')
            }
        }
    }, [history, hasChangedFields])

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
                        value={currentCodeMonitorState.description}
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
                <select className="form-control my-2 code-monitor-form__owner-dropdown w-auto" disabled={true}>
                    <option value={authenticatedUser.displayName || authenticatedUser.username}>
                        {authenticatedUser.username}
                    </option>
                </select>
                <small className="text-muted">Event history and configuration will not be shared.</small>
            </div>
            <hr className="code-monitor-form__horizontal-rule" />
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
            <hr className="code-monitor-form__horizontal-rule" />
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
                            !formCompletion.triggerCompleted ||
                            codeMonitorOrError === LOADING ||
                            !hasChangedFields
                        }
                        className="btn btn-primary mr-2 test-submit-monitor"
                    >
                        {submitButtonLabel}
                    </button>
                    <button type="button" className="btn btn-outline-secondary test-cancel-monitor" onClick={onCancel}>
                        {/* TODO: this should link somewhere */}
                        Cancel
                    </button>
                </div>
                {isErrorLike(codeMonitorOrError) && (
                    <div className="alert alert-danger">Failed to create monitor: {codeMonitorOrError.message}</div>
                )}
            </div>
        </Form>
    )
}
