import * as React from 'react'
import * as H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { ThemeProps } from '../../../../../shared/src/theme'
import classNames from 'classnames'
import { MonacoSettingsEditor } from '../../../settings/MonacoSettingsEditor'
import { ActionJob } from './ActionJob'
import SyncIcon from 'mdi-react/SyncIcon'
import { fetchActionExecutionByID, cancelActionExecution as _cancelActionExecution } from './backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { merge, of, Subject } from 'rxjs'
import { switchMap, repeatWhen, delay, catchError } from 'rxjs/operators'
import { Link } from '../../../../../shared/src/components/Link'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import CollapseAllIcon from 'mdi-react/CollapseAllIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { formatDistance, parseISO } from 'date-fns'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import { ErrorAlert } from '../../../components/alerts'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { asError, isErrorLike } from '../../../../../shared/src/util/errors'

interface Props extends ThemeProps {
    actionExecutionID: string

    location: H.Location
    history: H.History
}

export const ActionExecution: React.FunctionComponent<Props> = ({
    actionExecutionID,
    isLightTheme,
    history,
    location,
}) => {
    const [alertError, setAlertError] = React.useState<Error | undefined>()
    const executionUpdates = React.useMemo(() => new Subject<void>(), [])
    const nextExecutionUpdate = React.useCallback(() => executionUpdates.next(), [executionUpdates])
    const execution = useObservable(
        React.useMemo(
            () =>
                merge(of(undefined), executionUpdates).pipe(
                    switchMap(() =>
                        fetchActionExecutionByID(actionExecutionID).pipe(
                            catchError(error => {
                                setAlertError(error)
                                return []
                            }),
                            // refresh every 2s after the previous request finished so it feels more 'alive'
                            repeatWhen(obs => obs.pipe(delay(2000)))
                        )
                    )
                ),
            [actionExecutionID, executionUpdates]
        )
    )
    const [isCanceling, setIsCanceling] = React.useState<boolean | Error>(false)
    const cancelActionExecution: React.MouseEventHandler = async () => {
        setIsCanceling(true)
        try {
            await _cancelActionExecution(actionExecutionID)
            nextExecutionUpdate()
        } catch (error) {
            setIsCanceling(asError(error))
        }
    }
    if (execution === undefined) {
        return <LoadingSpinner />
    }
    if (execution === null) {
        return <h3>Execution not found!</h3>
    }
    return (
        <>
            <PageTitle title="Action execution #19292" />
            <h1 className={classNames(isLightTheme && 'text-info')}>
                Running action #19292
                {/* TODO: this is copied from ActionExecutionNode, extract to separate component */}
                {execution.status.state === GQL.BackgroundProcessState.COMPLETED && (
                    <CheckboxBlankCircleIcon
                        data-tooltip="Execution has finished successful"
                        className="icon-inline ml-3 text-success"
                    />
                )}
                {execution.status.state === GQL.BackgroundProcessState.PROCESSING && (
                    <SyncIcon
                        data-tooltip="Execution is running"
                        className="icon-inline ml-3 text-info icon-spinning"
                    />
                )}
                {execution.status.state === GQL.BackgroundProcessState.CANCELED && (
                    <CollapseAllIcon
                        data-tooltip="Execution has been canceled"
                        className="icon-inline ml-3 text-warning"
                    />
                )}
                {execution.status.state === GQL.BackgroundProcessState.ERRORED && (
                    <AlertCircleIcon data-tooltip="Execution has failed" className="icon-inline ml-3 text-danger" />
                )}
            </h1>
            {alertError && <ErrorAlert error={alertError} />}
            {execution.invocationReason === GQL.ActionExecutionInvocationReason.SCHEDULE && (
                <div className="alert alert-info">
                    This execution is part of a scheduled action.
                    <br />
                    <code>{execution.action.schedule}</code>
                </div>
            )}
            {execution.invocationReason === GQL.ActionExecutionInvocationReason.SAVED_SEARCH && (
                <div className="alert alert-info">
                    This execution is run because the results of saved search "
                    <a href="">
                        <i>{execution.action.savedSearch?.description}</i>
                    </a>
                    " changed.
                </div>
            )}
            <h2>Action definition</h2>
            <MonacoSettingsEditor
                isLightTheme={isLightTheme}
                readOnly={true}
                language="json"
                value={execution.definition.steps}
                height={200}
                className="mb-3"
            />
            <h2>Action status</h2>
            <div>
                <div className="alert alert-info mt-1">
                    <h3>Want faster execution? To add more agents:</h3>
                    <p>Use the below token to register your agent to this Sourcegraph instance</p>
                    <input className="form-control mb-2" readOnly={true} value="KK3DK99AA1291S8" />
                    <div>
                        <code>
                            src actions runner --sourcegraph-url={window.location.protocol}//{window.location.host}{' '}
                            --token=KK3DK99AA1291S8
                        </code>
                    </div>
                    or
                    <div>
                        <code>
                            kubectl create secret generic sg-token --from-literal=token=KK3DK99AA1291S8
                            --from-literal=sourcegraphUrl={window.location.protocol}//{window.location.host}
                            <br />
                            kubectl apply -f {window.location.protocol}//{window.location.host}
                            /.api/runner-kubeconfig.yml
                        </code>
                    </div>
                </div>
                {execution.status.state === GQL.BackgroundProcessState.PROCESSING && (
                    <div className="alert alert-info d-flex justify-content-between align-items-center">
                        <p>
                            {execution.executionStart ? (
                                <>
                                    <SyncIcon className="icon-inline icon-spinning" /> Execution is running since{' '}
                                    {formatDistance(parseISO(execution.executionStart), new Date())}.
                                </>
                            ) : (
                                <>Execution is awaiting an agent to pick up jobs.</>
                            )}
                        </p>
                        <button
                            type="button"
                            className="btn btn-danger"
                            disabled={isCanceling === true}
                            onClick={cancelActionExecution}
                        >
                            {isErrorLike(isCanceling) && <AlertCircleIcon data-tooltip={isCanceling.message} />} Cancel
                        </button>
                    </div>
                )}
                {execution.status.state === GQL.BackgroundProcessState.COMPLETED && (
                    <p>
                        <CheckCircleIcon className="icon-inline text-success" /> Execution finished
                        {execution.executionEnd && (
                            <> {formatDistance(parseISO(execution.executionEnd), new Date())} ago</>
                        )}
                        .
                    </p>
                )}
                {execution.status.state === GQL.BackgroundProcessState.CANCELED && <p>Execution has been canceled.</p>}
                {execution.status.state === GQL.BackgroundProcessState.ERRORED && <p>Execution has errored.</p>}
            </div>
            {execution.status.state === GQL.BackgroundProcessState.PROCESSING && (
                <>
                    <div className="progress">
                        <div
                            className="progress-bar"
                            /* eslint-disable-next-line react/forbid-dom-props */
                            style={{
                                width:
                                    (execution.status.completedCount /
                                        (execution.status.pendingCount + execution.status.completedCount)) *
                                        100 +
                                    '%',
                            }}
                        >
                            &nbsp;
                        </div>
                    </div>
                    <p className="text-center mt-1 w-100">
                        {execution.status.completedCount} /{' '}
                        {execution.status.pendingCount + execution.status.completedCount} jobs ran
                    </p>
                </>
            )}
            <h2>Action jobs</h2>
            <ul className="list-group mb-3">
                {execution.jobs.nodes.map(actionJob => (
                    <ActionJob
                        isLightTheme={isLightTheme}
                        actionJob={actionJob}
                        key={actionJob.id}
                        history={history}
                        location={location}
                        onRetry={nextExecutionUpdate}
                    />
                ))}
            </ul>
            <h2>Action summary</h2>
            <p>
                {[GQL.BackgroundProcessState.ERRORED, GQL.BackgroundProcessState.COMPLETED].includes(
                    execution.status.state
                ) && (
                    <>
                        The action execution finished
                        {execution.status.state === GQL.BackgroundProcessState.ERRORED
                            ? ', but some tasks failed'
                            : ' successfully'}
                        .
                    </>
                )}{' '}
                {execution.status.state === GQL.BackgroundProcessState.CANCELED && (
                    <>The action execution has been canceled.</>
                )}{' '}
                {execution.status.state === GQL.BackgroundProcessState.PROCESSING && (
                    <>The action execution is still running.</>
                )}{' '}
                <span className="badge badge-secondary">{execution.jobs.nodes.filter(job => job.diff).length}</span>{' '}
                patches have been created{execution.status.state === GQL.BackgroundProcessState.PROCESSING && ' so far'}
                .
            </p>
            {execution.action.campaign ? (
                <>
                    <p className="alert alert-info">
                        This action belongs to campaign{' '}
                        <Link to={`/campaigns/${execution.action.campaign.id}`}>
                            <em>{execution.action.campaign.name}</em>
                        </Link>
                        . The patches are automatically applied to the campaign upon completion.
                    </p>
                </>
            ) : execution.patchSet ? (
                <div className="mb-3">
                    <Link to={`/campaigns/new?plan=${execution.patchSet.id}`} className="btn btn-primary">
                        Create campaign
                    </Link>
                    <Link to={`/campaigns/update?plan=${execution.patchSet.id}`} className="btn btn-primary ml-2">
                        Update existing campaign
                    </Link>
                </div>
            ) : (
                <p>
                    Please stand by as we generate the patch set. Creating and updating a campaign will be possible
                    soon.
                </p>
            )}
        </>
    )
}
