import * as React from 'react'
import * as H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../shared/src/theme'
import { Collapsible } from '../../../components/Collapsible'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import { parseISO, formatDistance } from 'date-fns/esm'
import { DiffStat } from '../../../components/diff/DiffStat'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CollapseAllIcon from 'mdi-react/CollapseAllIcon'
import { FileDiffNode } from '../../../components/diff/FileDiffNode'
import { retryActionJob } from './backend'
import { asError } from '../../../../../shared/src/util/errors'

interface Props extends ThemeProps {
    actionJob: GQL.IActionJob
    location: H.Location
    history: H.History
    onRetry: () => void
}

export const ActionJob: React.FunctionComponent<Props> = ({ isLightTheme, actionJob, location, history, onRetry }) => {
    const [isRetrying, setIsRetrying] = React.useState<boolean>(false)
    const [retryError, setRetryError] = React.useState<Error | undefined>()
    const retry = React.useCallback(async () => {
        setIsRetrying(true)
        setRetryError(undefined)
        try {
            await new Promise(resolve => setTimeout(resolve, 4000))
            await retryActionJob(actionJob.id)
            onRetry()
        } catch (error) {
            setRetryError(asError(error))
        } finally {
            setIsRetrying(false)
        }
    }, [actionJob.id, onRetry])
    return (
        <>
            <li className="list-group-item">
                <Collapsible
                    title={
                        <div className="ml-2 d-flex justify-content-between align-content-center">
                            <div className="flex-grow-1">
                                <h3 className="mb-1">Run on {actionJob.repository.name}</h3>
                                <p className="mb-0">
                                    {actionJob.runner ? (
                                        <small className="text-monospace">Runner {actionJob.runner.name}</small>
                                    ) : (
                                        <i>Awaiting runner assignment</i>
                                    )}
                                </p>
                            </div>
                            {actionJob.executionStart && !actionJob.executionEnd && (
                                <div className="flex-grow-0">
                                    <p className="m-0 text-right mr-2">
                                        Started {formatDistance(parseISO(actionJob.executionStart), new Date())} ago
                                    </p>
                                </div>
                            )}
                            {actionJob.executionEnd && (
                                <div className="flex-grow-0">
                                    <p className="m-0 text-right mr-2">
                                        {actionJob.state === GQL.ActionJobState.ERRORED ? 'Failed' : 'Finished'}{' '}
                                        {formatDistance(parseISO(actionJob.executionEnd), new Date())} ago
                                    </p>
                                </div>
                            )}
                            <div className="flex-grow-0">
                                {actionJob.state === GQL.ActionJobState.COMPLETED && (
                                    <div className="d-flex justify-content-end">
                                        <CheckboxBlankCircleIcon
                                            data-tooltip="Task is running"
                                            className="text-success"
                                        />
                                    </div>
                                )}
                                {actionJob.state === GQL.ActionJobState.PENDING && (
                                    <div className="d-flex justify-content-end">
                                        <CheckboxBlankCircleIcon
                                            data-tooltip="Task is pending"
                                            className="text-warning"
                                        />
                                    </div>
                                )}
                                {actionJob.state === GQL.ActionJobState.RUNNING && (
                                    <div className="d-flex justify-content-end">
                                        <SyncIcon data-tooltip="Task is running" className="text-info" />
                                    </div>
                                )}
                                {actionJob.state === GQL.ActionJobState.CANCELED && (
                                    <div className="d-flex justify-content-end">
                                        <CollapseAllIcon data-tooltip="Task was canceled" className="text-warn" />
                                    </div>
                                )}
                                {actionJob.state === GQL.ActionJobState.ERRORED && (
                                    <>
                                        <div className="d-flex justify-content-end">
                                            <AlertCircleIcon data-tooltip="Task has failed" className="text-danger" />
                                        </div>
                                        <button
                                            type="button"
                                            className="btn btn-sm btn-secondary"
                                            disabled={isRetrying}
                                            onClick={retry}
                                        >
                                            Retry
                                            {retryError && (
                                                <AlertCircleIcon
                                                    className="ml-2 icon-inline text-danger"
                                                    data-tooltip={retryError}
                                                />
                                            )}
                                        </button>
                                    </>
                                )}
                                {actionJob.diff?.fileDiffs.diffStat && (
                                    <DiffStat {...actionJob.diff.fileDiffs.diffStat} />
                                )}
                            </div>
                        </div>
                    }
                    titleClassName="flex-grow-1"
                    wholeTitleClickable={false}
                >
                    {actionJob.log && (
                        <>
                            {' '}
                            <h5 className="mb-1">Log output</h5>
                            <div
                                className="p-1 mb-3"
                                // eslint-disable-next-line react/forbid-dom-props
                                style={{
                                    border: '1px solid grey',
                                    background: 'black',
                                    color: '#fff',
                                    overflowX: 'auto',
                                    maxHeight: '200px',
                                }}
                            >
                                <code dangerouslySetInnerHTML={{ __html: actionJob.log }}></code>
                                {actionJob.state === GQL.ActionJobState.RUNNING && (
                                    <div>
                                        <SyncIcon className="icon-inline" />
                                    </div>
                                )}
                            </div>
                        </>
                    )}
                    <h5 className="mb-1">Generated diff</h5>
                    {/* eslint-disable-next-line react/forbid-dom-props */}
                    <div className="p-1" style={{ border: '1px solid grey' }}>
                        {actionJob.diff ? (
                            actionJob.diff.fileDiffs.nodes.map(fileDiffNode => (
                                <FileDiffNode
                                    isLightTheme={isLightTheme}
                                    node={fileDiffNode}
                                    lineNumbers={true}
                                    location={location}
                                    history={history}
                                    persistLines={false}
                                    // todo: is this a good key?
                                    key={fileDiffNode.internalID}
                                ></FileDiffNode>
                            ))
                        ) : (
                            <p className="text-muted">No diff has been generated</p>
                        )}
                    </div>
                </Collapsible>
            </li>
        </>
    )
}
