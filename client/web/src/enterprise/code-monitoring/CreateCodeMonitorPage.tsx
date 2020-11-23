import * as H from 'history'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Form } from '../../../../branded/src/components/Form'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'
import classnames from 'classnames'
import { useEventObservable } from '../../../../shared/src/util/useObservable'
import { createCodeMonitor } from './backend'
import { MonitorEmailPriority } from '../../../../shared/src/graphql/schema'
import { Observable } from 'rxjs'
import { catchError, mergeMap, startWith, tap } from 'rxjs/operators'
import { asError, isErrorLike } from '../../../../shared/src/util/errors'
import { Link } from '../../../../shared/src/components/Link'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { deriveInputClassName, useInputValidation } from '../../../../shared/src/util/useInputValidation'
import { scanSearchQuery } from '../../../../shared/src/search/parser/scanner'
import { FilterType } from '../../../../shared/src/search/interactive/util'
import { resolveFilter, validateFilter } from '../../../../shared/src/search/parser/filters'

export interface CreateCodeMonitorPageProps extends BreadcrumbsProps, BreadcrumbSetters {
    location: H.Location
    authenticatedUser: AuthenticatedUser
}

interface FormCompletionSteps {
    triggerCompleted: boolean
    actionCompleted: boolean
}

interface Action {
    recipient: string
    enabled: boolean
}
interface CodeMonitorFields {
    description: string
    query: string
    enabled: boolean
    action: Action
}

export const CreateCodeMonitorPage: React.FunctionComponent<CreateCodeMonitorPageProps> = props => {
    props.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Create Code Monitor',
                element: <>Create new code monitor</>,
            }),
            []
        )
    )

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
        <div className="container mt-3 web-content">
            <PageTitle title="Create new code monitor" />
            <PageHeader title="Create new code monitor" icon={VideoInputAntennaIcon} />
            Code monitors watch your code for specific triggers and run actions in response.{' '}
            <a href="" target="_blank" rel="noopener">
                {/* TODO: populate link */}
                Learn more
            </a>
            <Form className="my-4" onSubmit={createRequest}>
                <div className="flex mb-4">
                    Name
                    <div>
                        <input
                            type="text"
                            className="form-control my-2"
                            required={true}
                            onChange={event => {
                                onNameChange(event.target.value)
                            }}
                            autoFocus={true}
                        />
                    </div>
                    <small className="text-muted">
                        Give it a short, descriptive name to reference events on Sourcegraph and in notifications. Do
                        not include:{' '}
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
                    <TriggerArea
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
                    <ActionArea
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
        </div>
    )
}

interface TriggerAreaProps {
    query: string
    onQueryChange: (query: string) => void
    triggerCompleted: boolean
    setTriggerCompleted: () => void
}

const isDiffOrCommit = (value: string): boolean => value === 'diff' || value === 'commit'

const TriggerArea: React.FunctionComponent<TriggerAreaProps> = ({
    query,
    onQueryChange,
    triggerCompleted,
    setTriggerCompleted,
}) => {
    const [showQueryForm, setShowQueryForm] = useState(false)
    const toggleQueryForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowQueryForm(show => !show)
    }, [])

    const editOrCompleteForm: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            toggleQueryForm(event)
            setTriggerCompleted()
        },
        [setTriggerCompleted, toggleQueryForm]
    )

    const [queryState, nextQueryFieldChange, queryInputReference] = useInputValidation(
        useMemo(
            () => ({
                synchronousValidators: [
                    (value: string) => {
                        const tokens = scanSearchQuery(value)
                        if (tokens.type === 'success') {
                            const filters = tokens.term.filter(token => token.type === 'filter')
                            const hasTypeDiffOrCommitFilter = filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.type &&
                                    ((filter.value?.type === 'literal' &&
                                        filter.value &&
                                        isDiffOrCommit(filter.value.value)) ||
                                        (filter.value?.type === 'quoted' &&
                                            filter.value &&
                                            isDiffOrCommit(filter.value.quotedValue)))
                            )
                            const hasPatternTypeFilter = filters.some(
                                filter =>
                                    filter.type === 'filter' &&
                                    resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                                    filter.value &&
                                    validateFilter(filter.field.value, filter.value)
                            )
                            if (hasTypeDiffOrCommitFilter && hasPatternTypeFilter) {
                                return undefined
                            }
                            if (!hasTypeDiffOrCommitFilter) {
                                return 'Code monitors require queries to specify either `type:commit` or `type:diff`.'
                            }
                            if (!hasPatternTypeFilter) {
                                return 'Code monitors require queries to specify a `patternType:` of literal, regexp, or structural.'
                            }
                        }
                        return 'Failed to parse query'
                    },
                ],
            }),
            []
        )
    )

    useEffect(() => {
        if (queryState.kind === 'VALID') {
            onQueryChange(queryState.value)
        }
    }, [onQueryChange, queryState])

    return (
        <>
            <h3>Trigger</h3>
            <div className="card p-3 my-3">
                {!showQueryForm && !triggerCompleted && (
                    <>
                        <button
                            type="button"
                            onClick={toggleQueryForm}
                            className="btn btn-link font-weight-bold p-0 text-left test-trigger-button"
                        >
                            When there are new search results
                        </button>
                        <span className="text-muted">
                            This trigger will fire when new search results are found for a given search query.
                        </span>
                    </>
                )}
                {showQueryForm && (
                    <>
                        <div className="font-weight-bold">When there are new search results</div>
                        <span className="text-muted">
                            This trigger will fire when new search results are found for a given search query.
                        </span>
                        <div className="create-monitor-page__query-input">
                            <input
                                type="text"
                                className={classnames(
                                    'create-monitor-page__query-input-field form-control my-2 test-trigger-input',
                                    deriveInputClassName(queryState)
                                )}
                                onChange={nextQueryFieldChange}
                                value={queryState.value}
                                required={true}
                                autoFocus={true}
                                ref={queryInputReference}
                            />
                            {queryState.kind === 'VALID' && (
                                <Link
                                    to={buildSearchURLQuery(query, SearchPatternType.literal, false)}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="create-monitor-page__query-input-preview-link"
                                >
                                    Preview results <OpenInNewIcon />
                                </Link>
                            )}
                            {queryState.kind === 'INVALID' && (
                                <small className="invalid-feedback mb-4">{queryState.reason}</small>
                            )}
                            {(queryState.kind === 'NOT_VALIDATED' || queryState.kind === 'VALID') && (
                                <div className="d-flex mb-4 flex-column">
                                    <small className="text-muted">
                                        Code monitors only support <code className="bg-code">type:diff</code> and{' '}
                                        <code className="bg-code">type:commit</code> search queries.
                                    </small>
                                </div>
                            )}
                        </div>
                        <div>
                            <button
                                className="btn btn-outline-secondary mr-1 test-submit-trigger"
                                onClick={editOrCompleteForm}
                                onSubmit={editOrCompleteForm}
                                type="submit"
                                disabled={queryState.kind !== 'VALID'}
                            >
                                Continue
                            </button>
                            <button type="button" className="btn btn-outline-secondary" onClick={editOrCompleteForm}>
                                Cancel
                            </button>
                        </div>
                    </>
                )}
                {triggerCompleted && (
                    <div className="d-flex justify-content-between align-items-center">
                        <div>
                            <div className="font-weight-bold">When there are new search results</div>
                            <code className="text-muted">{query}</code>
                        </div>
                        <div>
                            <button type="button" onClick={editOrCompleteForm} className="btn btn-link p-0 text-left">
                                Edit
                            </button>
                        </div>
                    </div>
                )}
            </div>
            <small className="text-muted">
                {' '}
                What other events would you like to monitor? {/* TODO: populate link */}
                <a href="" target="_blank" rel="noopener">
                    {/* TODO: populate link */}
                    Share feedback.
                </a>
            </small>
        </>
    )
}

interface ActionAreaProps {
    actionCompleted: boolean
    setActionCompleted: () => void
    disabled: boolean
    authenticatedUser: AuthenticatedUser
    onActionChange: (action: Action) => void
}

const ActionArea: React.FunctionComponent<ActionAreaProps> = ({
    actionCompleted,
    setActionCompleted,
    disabled,
    authenticatedUser,
    onActionChange,
}) => {
    const [showEmailNotificationForm, setShowEmailNotificationForm] = useState(false)
    const toggleEmailNotificationForm: React.FormEventHandler = useCallback(event => {
        event.preventDefault()
        setShowEmailNotificationForm(show => !show)
    }, [])

    const editOrCompleteForm: React.FormEventHandler = useCallback(
        event => {
            event?.preventDefault()
            toggleEmailNotificationForm(event)
            setActionCompleted()
        },
        [toggleEmailNotificationForm, setActionCompleted]
    )

    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(true)
    const toggleEmailNotificationEnabled: (value: boolean) => void = useCallback(
        enabled => {
            setEmailNotificationEnabled(enabled)
            onActionChange({ recipient: authenticatedUser.email, enabled })
        },
        [authenticatedUser, onActionChange]
    )

    return (
        <>
            <h3 className="mb-1">Actions</h3>
            <span className="text-muted">Run any number of actions in response to an event</span>
            <div className="card p-3 my-3">
                {/* This should be its own component when you can add multiple email actions */}
                {!showEmailNotificationForm && !actionCompleted && (
                    <>
                        <button
                            type="button"
                            onClick={toggleEmailNotificationForm}
                            className="btn btn-link font-weight-bold p-0 text-left test-action-button"
                            disabled={disabled}
                        >
                            Send email notifications
                        </button>
                        <span className="text-muted">Deliver email notifications to specified recipients.</span>
                    </>
                )}
                {showEmailNotificationForm && !actionCompleted && (
                    <>
                        <div className="font-weight-bold">Send email notifications</div>
                        <span className="text-muted">Deliver email notifications to specified recipients.</span>
                        <div className="mt-4">
                            Recipients
                            <input
                                type="text"
                                className="form-control my-2"
                                value={`${authenticatedUser.email || ''} (you)`}
                                disabled={true}
                                autoFocus={true}
                            />
                            <small className="text-muted">
                                Code monitors are currently limited to sending emails to your primary email address.
                            </small>
                        </div>
                        <div className="flex my-4">
                            <Toggle
                                title="Enabled"
                                value={emailNotificationEnabled}
                                onToggle={toggleEmailNotificationEnabled}
                                className="mr-2"
                            />
                            Enabled
                        </div>
                        <div>
                            <button
                                type="submit"
                                className="btn btn-outline-secondary mr-1 test-submit-action"
                                onClick={editOrCompleteForm}
                                onSubmit={editOrCompleteForm}
                            >
                                Done
                            </button>
                            <button type="button" className="btn btn-outline-secondary" onClick={editOrCompleteForm}>
                                Cancel
                            </button>
                        </div>
                    </>
                )}
                {actionCompleted && (
                    <div className="d-flex justify-content-between align-items-center">
                        <div>
                            <div className="font-weight-bold">Send email notifications</div>
                            <span className="text-muted">{authenticatedUser.email}</span>
                        </div>
                        <div className="d-flex">
                            <div className="flex my-4">
                                <Toggle
                                    title="Enabled"
                                    value={emailNotificationEnabled}
                                    onToggle={toggleEmailNotificationEnabled}
                                    className="mr-2"
                                />
                            </div>
                            <button type="button" onClick={editOrCompleteForm} className="btn btn-link p-0 text-left">
                                Edit
                            </button>
                        </div>
                    </div>
                )}
            </div>
            <small className="text-muted">
                What other actions would you like to do?{' '}
                <a href="" target="_blank" rel="noopener">
                    {/* TODO: populate link */}
                    Share feedback.
                </a>
            </small>
        </>
    )
}
