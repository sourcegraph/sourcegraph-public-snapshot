import React from 'react'
import * as H from 'history'
import { PageTitle } from '../../../components/PageTitle'
import { ThemeProps } from '../../../../../shared/src/theme'
import classNames from 'classnames'
import { MonacoSettingsEditor } from '../../../settings/MonacoSettingsEditor'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { merge, of, Subject } from 'rxjs'
import { switchMap, filter, tap } from 'rxjs/operators'
import { fetchActionByID, createActionExecution, createAction, updateAction } from './backend'
import { ActionExecutionNode } from './list/ActionExecutionNode'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { Form } from '../../../components/Form'

interface Props extends ThemeProps {
    actionID?: string
    history: H.History
}

export const Action: React.FunctionComponent<Props> = ({ actionID, isLightTheme, history }) => {
    const [name, setName] = React.useState<string>('')
    const [steps, setSteps] = React.useState<string>('')
    const [isLoading, setIsLoading] = React.useState<boolean>()
    const nextUpdate = React.useMemo(() => new Subject<void>(), [])
    const action = useObservable(
        React.useMemo(
            () =>
                merge(nextUpdate, of(undefined).pipe(filter(() => !!actionID))).pipe(
                    switchMap(() => fetchActionByID(actionID!)),
                    tap(action => {
                        if (action) {
                            setName(action.name)
                            setSteps(action.definition.steps)
                        }
                    })
                ),
            [actionID, nextUpdate]
        )
    )
    const onSubmit = React.useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                if (actionID) {
                    await updateAction(actionID, steps ?? '')
                } else {
                    const action = await createAction({ definition: steps, name })
                    history.push('/campaigns/actions/' + action.id)
                }
                nextUpdate.next()
            } finally {
                setIsLoading(false)
            }
        },
        [steps, name, actionID, history, nextUpdate]
    )
    const createExecution = React.useCallback(async () => {
        if (action) {
            await createActionExecution(action.id)
            nextUpdate.next()
        }
    }, [action, nextUpdate])
    if (actionID && action === undefined) {
        return <LoadingSpinner />
    }
    if (actionID && action === null) {
        return <h3>Action not found!</h3>
    }
    return (
        <>
            <PageTitle title={action ? `Action "${action.name}"` : 'New action'} />
            {action ? (
                <h1 className={classNames(isLightTheme && 'text-info')}>Action {action.name}</h1>
            ) : (
                <h1 className={classNames(isLightTheme && 'text-info')}>Create new action</h1>
            )}
            {action?.schedule && (
                <div className="alert alert-info">
                    This action is a scheduled action.
                    <br />
                    <code>{action.schedule}</code>
                </div>
            )}
            {action?.savedSearch && (
                <div className="alert alert-info">
                    This action executes whenever the results of saved search "
                    <a href="">
                        <i>{action.savedSearch?.description}</i>
                    </a>
                    " change.
                </div>
            )}
            <Form onSubmit={onSubmit}>
                <div className="form-group">
                    <h2>
                        <label htmlFor="actionName">Name</label>
                    </h2>
                    <input
                        type="text"
                        id="actionName"
                        name="name"
                        value={name}
                        onChange={e => setName(e.target.value)}
                        className="form-control"
                        disabled={isLoading}
                        required={true}
                    />
                </div>
                <h2>Action definition</h2>
                <MonacoSettingsEditor
                    isLightTheme={isLightTheme}
                    readOnly={isLoading}
                    language="json"
                    value={steps}
                    onChange={setSteps}
                    height={200}
                    className="mb-3"
                />
                {!action && (
                    <button className="btn btn-primary mb-3" type="submit" disabled={isLoading}>
                        Create action
                    </button>
                )}
                {action && (
                    <button className="btn btn-primary mb-3" type="submit" disabled={isLoading}>
                        Update action definition
                    </button>
                )}
            </Form>
            {action && (
                <>
                    <h2>Action executions</h2>
                    <button className="btn btn-primary mb-3" type="button" onClick={createExecution}>
                        Create new execution
                    </button>
                    <ul className="list-group mb-3">
                        {action.actionExecutions.nodes.map(actionExecution => (
                            <ActionExecutionNode node={actionExecution} key={actionExecution.id} />
                        ))}
                        {action.actionExecutions.totalCount === 0 && (
                            <p className="text-muted">No executions were run yet.</p>
                        )}
                    </ul>
                </>
            )}
        </>
    )
}
