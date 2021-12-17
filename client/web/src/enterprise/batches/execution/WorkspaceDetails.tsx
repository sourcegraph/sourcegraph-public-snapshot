import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import ContentSaveIcon from 'mdi-react/ContentSaveIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import LinkVariantRemoveIcon from 'mdi-react/LinkVariantRemoveIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { useHistory } from 'react-router'
import { delay, repeatWhen } from 'rxjs/operators'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Badge, LoadingSpinner, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../components/alerts'
import { Collapsible } from '../../../components/Collapsible'
import { DiffStat } from '../../../components/diff/DiffStat'
import { FileDiffConnection } from '../../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../../components/diff/FileDiffNode'
import { FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import { HeroPage } from '../../../components/HeroPage'
import { LogOutput } from '../../../components/LogOutput'
import { Duration } from '../../../components/time/Duration'
import {
    BatchSpecWorkspaceChangesetSpecFields,
    BatchSpecWorkspaceState,
    BatchSpecWorkspaceStepFields,
    Scalars,
} from '../../../graphql-operations'
import { queryChangesetSpecFileDiffs } from '../preview/list/backend'
import { ChangesetSpecFileDiffConnection } from '../preview/list/ChangesetSpecFileDiffConnection'

import { fetchBatchSpecWorkspace, queryBatchSpecWorkspaceStepFileDiffs, retryWorkspaceExecution } from './backend'
import { TimelineModal } from './TimelineModal'
import styles from './WorkspaceDetails.module.scss'
import { WorkspaceStateIcon } from './WorkspaceStateIcon'

export interface WorkspaceDetailsProps extends ThemeProps {
    id: Scalars['ID']
}

export const WorkspaceDetails: React.FunctionComponent<WorkspaceDetailsProps> = ({ id, isLightTheme }) => {
    const history = useHistory()
    const onClose = useCallback(() => {
        history.push(history.location.pathname)
    }, [history])

    // Fetch and poll latest workspace information.
    const workspace = useObservable(
        useMemo(() => fetchBatchSpecWorkspace(id).pipe(repeatWhen(notifier => notifier.pipe(delay(2500)))), [id])
    )

    const [retrying, setRetrying] = useState<boolean | ErrorLike>(false)
    const onRetry = useCallback(async () => {
        setRetrying(true)
        try {
            await retryWorkspaceExecution(id)
            setRetrying(false)
        } catch (error) {
            setRetrying(asError(error))
        }
    }, [id])

    const [showTimeline, setShowTimeline] = useState<boolean>(false)
    const toggleShowTimeline = useCallback(() => {
        setShowTimeline(true)
    }, [])
    const onDismissTimeline = useCallback(() => {
        setShowTimeline(false)
    }, [])

    if (workspace === undefined) {
        return <LoadingSpinner />
    }

    if (workspace === null) {
        return <NotFoundPage />
    }

    if (isErrorLike(workspace)) {
        return <ErrorAlert error={workspace} />
    }

    return (
        <>
            {showTimeline && <TimelineModal node={workspace} onCancel={onDismissTimeline} />}
            <div className="d-flex justify-content-between">
                <h3>
                    <WorkspaceStateIcon cachedResultFound={workspace.cachedResultFound} state={workspace.state} />{' '}
                    {workspace.repository.name}{' '}
                    <Link to={workspace.repository.url}>
                        <ExternalLinkIcon className="icon-inline" />
                    </Link>
                </h3>
                <button type="button" className="btn btn-link btn-sm p-0 ml-2" onClick={onClose}>
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
            <div className="text-muted">
                {workspace.path && <>{workspace.path} | </>}
                <SourceBranchIcon className="icon-inline" /> base: <strong>{workspace.branch.abbrevName}</strong>
                {workspace.startedAt && (
                    <>
                        {' '}
                        | Total time:{' '}
                        <strong>
                            <Duration start={workspace.startedAt} end={workspace.finishedAt ?? undefined} />
                        </strong>
                    </>
                )}
                {typeof workspace.placeInQueue === 'number' && (
                    <>
                        {' '}
                        | <SyncIcon className="icon-inline" />{' '}
                        <strong>
                            <NumberInQueue number={workspace.placeInQueue} />
                        </strong>{' '}
                        in queue
                    </>
                )}
                {!workspace.cachedResultFound && workspace.state !== BatchSpecWorkspaceState.SKIPPED && (
                    <>
                        {' '}
                        |{' '}
                        <button type="button" className="text-muted btn btn-link m-0 p-0" onClick={toggleShowTimeline}>
                            Timeline
                        </button>
                    </>
                )}
            </div>
            <hr />
            {workspace.failureMessage && (
                <>
                    <div className="d-flex my-3 w-100">
                        <ErrorAlert error={workspace.failureMessage} className="flex-grow-1 mb-0" />
                        <button
                            type="button"
                            className="btn btn-outline-danger ml-2"
                            onClick={onRetry}
                            disabled={retrying === true}
                        >
                            <SyncIcon className="icon-inline" /> Retry
                        </button>
                    </div>
                    {isErrorLike(retrying) && <ErrorAlert error={retrying} />}
                </>
            )}
            {workspace.state === BatchSpecWorkspaceState.SKIPPED && workspace.ignored && (
                <p className="text-muted text-center py-3 mb-0">
                    <strong>
                        <LinkVariantRemoveIcon className="icon-inline" /> This workspace has been skipped because a{' '}
                        <code>.batchignore</code> file was found.
                    </strong>
                </p>
            )}
            {workspace.state === BatchSpecWorkspaceState.SKIPPED && workspace.unsupported && (
                <p className="text-muted text-center py-3 mb-0">
                    <strong>
                        <LinkVariantRemoveIcon className="icon-inline" /> This workspace has been skipped because it is
                        on an unsupported code host.
                    </strong>
                </p>
            )}
            {workspace.changesetSpecs && workspace.state === BatchSpecWorkspaceState.COMPLETED && (
                <div className="my-3">
                    {workspace.changesetSpecs.length === 0 && (
                        <p className="mb-0 text-muted">This workspace generated no changeset specs.</p>
                    )}
                    {workspace.changesetSpecs.map((changesetSpec, index) => (
                        <>
                            <ChangesetSpecNode
                                key={changesetSpec.id}
                                node={changesetSpec}
                                index={index}
                                isLightTheme={isLightTheme}
                            />
                            {index !== workspace.changesetSpecs!.length - 1 && <hr className="m-0" />}
                        </>
                    ))}
                </div>
            )}
            {workspace.steps.map((step, index) => (
                <React.Fragment key={step.number}>
                    <WorkspaceStep
                        step={step}
                        cachedResultFound={workspace.cachedResultFound}
                        workspaceID={workspace.id}
                        isLightTheme={isLightTheme}
                    />
                    {index !== workspace.steps.length - 1 && <hr className="m-0" />}
                </React.Fragment>
            ))}
        </>
    )
}

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

const NumberInQueue: React.FunctionComponent<{ number: number }> = ({ number }) => {
    let suffix: string
    console.log('NumberInQueue', number, number % 10)
    switch (number % 10) {
        case 1:
            suffix = 'st'
            break
        case 2:
            suffix = 'nd'
            break
        case 3:
            suffix = 'rd'
            break
        default:
            suffix = 'th'
    }
    return (
        <>
            {number}
            <sup>{suffix}</sup>
        </>
    )
}

const ChangesetSpecNode: React.FunctionComponent<
    { node: BatchSpecWorkspaceChangesetSpecFields; index: number } & ThemeProps
> = ({ node, index, isLightTheme }) => {
    const history = useHistory()

    // TODO: This should not happen. When the workspace is visibile, the changeset spec should be visible as well.
    if (node.__typename === 'HiddenChangesetSpec') {
        return (
            <div className="card">
                <div className="card-body">
                    <h4>Changeset in a hidden repo</h4>
                </div>
            </div>
        )
    }

    // This should not happen.
    if (node.description.__typename === 'ExistingChangesetReference') {
        return null
    }

    return (
        <Collapsible
            className="py-2"
            title={
                <div className="d-flex justify-content-between">
                    <div>
                        {' '}
                        <h4 className="mb-0 d-inline-block mr-2">
                            <strong>RESULT {index + 1}</strong>{' '}
                            {node.description.published !== null && (
                                <Badge className="text-uppercase">
                                    {publishBadgeLabel(node.description.published)}
                                </Badge>
                            )}{' '}
                        </h4>
                        <span className="text-muted">
                            <SourceBranchIcon className="icon-inline" />
                            changeset branch: <strong>{node.description.headRef}</strong>
                        </span>
                    </div>
                    <DiffStat {...node.description.diffStat} expandedCounts={true} />
                </div>
            }
            titleClassName="flex-grow-1"
            defaultExpanded={1 === 1}
        >
            <div className={classNames('card mt-2', styles.resultCard)}>
                <div className="card-body">
                    <h3>Changeset template</h3>
                    <h4>{node.description.title}</h4>
                    <p className="mb-0">{node.description.body}</p>
                    <p>
                        <strong>Published:</strong> <PublishedValue published={node.description.published} />
                    </p>
                    <Collapsible
                        title={<h3 className="mb-0">Changes</h3>}
                        titleClassName="flex-grow-1"
                        defaultExpanded={false}
                    >
                        <ChangesetSpecFileDiffConnection
                            history={history}
                            isLightTheme={isLightTheme}
                            location={history.location}
                            spec={node.id}
                            queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                        />
                    </Collapsible>
                </div>
            </div>
        </Collapsible>
    )
}

function publishBadgeLabel(state: Scalars['PublishedValue']): string {
    switch (state) {
        case 'draft':
            return 'will publish as draft'
        case false:
            return 'will not publish'
        case true:
            return 'will publish'
    }
}

const PublishedValue: React.FunctionComponent<{ published: Scalars['PublishedValue'] | null }> = ({ published }) => {
    if (published === null) {
        return <i>select from UI when applying</i>
    }
    if (published === 'draft') {
        return <>draft</>
    }
    return <>{String(published)}</>
}

interface WorkspaceStepProps extends ThemeProps {
    cachedResultFound: boolean
    step: BatchSpecWorkspaceStepFields
    workspaceID: Scalars['ID']
}

const WorkspaceStep: React.FunctionComponent<WorkspaceStepProps> = ({
    step,
    isLightTheme,
    workspaceID,
    cachedResultFound,
}) => {
    const outputLines = useMemo(() => {
        const outputLines = step.outputLines
        if (outputLines !== null) {
            if (
                outputLines.every(
                    line =>
                        line
                            .replaceAll(/'^std(out|err):'/g, '')
                            .replaceAll('\n', '')
                            .trim() === ''
                )
            ) {
                outputLines.push('stderr: This command did not produce any logs')
            }
            if (step.exitCode !== null) {
                outputLines.push(`\nstdout: \nstdout: Command exited with status ${step.exitCode}`)
            }
        }
        return outputLines
    }, [step.exitCode, step.outputLines])

    return (
        <Collapsible
            className="py-2"
            titleClassName="w-100"
            title={
                <div className="d-flex justify-content-between">
                    <div className={classNames('flex-grow-1', step.skipped && 'text-muted')}>
                        <StepStateIcon step={step} /> <strong>Step {step.number}</strong>{' '}
                        <span className="text-monospace text-ellipsis text-muted">{step.run}</span>
                    </div>
                    <div>{step.diffStat && <DiffStat {...step.diffStat} expandedCounts={true} />}</div>
                    <span className="text-monospace text-muted ml-2">
                        <StepTimer step={step} />
                    </span>
                </div>
            }
        >
            <div className={classNames('card mt-2', styles.stepCard)}>
                <div className="card-body">
                    {!step.skipped && (
                        <Tabs size="small" behavior="forceRender">
                            <TabList>
                                <Tab key="logs">Logs</Tab>
                                <Tab key="output-variables">Output variables</Tab>
                                <Tab key="diff">Diff</Tab>
                                <Tab key="files-env">Files / Env</Tab>
                                <Tab key="command-container">Commands / container</Tab>
                            </TabList>
                            <TabPanels>
                                <TabPanel key="logs">
                                    <div className="p-2">
                                        {!step.startedAt && <p className="text-muted mb-0">Step not started yet</p>}
                                        {step.startedAt && outputLines && <LogOutput text={outputLines.join('\n')} />}
                                    </div>
                                </TabPanel>
                                <TabPanel key="output-variables">
                                    <div className="p-2">
                                        {!step.startedAt && <p className="text-muted mb-0">Step not started yet</p>}
                                        {step.outputVariables?.length === 0 && (
                                            <p className="text-muted mb-0">No output variables specified</p>
                                        )}
                                        <ul className="mb-0">
                                            {step.outputVariables?.map(variable => (
                                                <li key={variable.name}>
                                                    {variable.name}: {variable.value}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                </TabPanel>
                                <TabPanel key="diff">
                                    <div className="p-2">
                                        {!step.startedAt && <p className="text-muted mb-0">Step not started yet</p>}
                                        {step.startedAt && (
                                            <WorkspaceStepFileDiffConnection
                                                isLightTheme={isLightTheme}
                                                step={step.number}
                                                workspaceID={workspaceID}
                                            />
                                        )}
                                    </div>
                                </TabPanel>
                                <TabPanel key="files-env">
                                    <div className="p-2">
                                        {step.environment.length === 0 && (
                                            <p className="text-muted mb-0">No environment variables specified</p>
                                        )}
                                        <ul className="mb-0">
                                            {step.environment.map(variable => (
                                                <li key={variable.name}>
                                                    {variable.name}: {variable.value}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                </TabPanel>
                                <TabPanel key="command-container">
                                    <div className="p-2 pb-0">
                                        {step.ifCondition !== null && (
                                            <>
                                                <h4>If condition</h4>
                                                <LogOutput text={step.ifCondition} className="mb-2" />
                                            </>
                                        )}
                                        <h4>Command</h4>
                                        <LogOutput text={step.run} className="mb-2" />
                                        <h4>Container</h4>
                                        <p className="text-monospace mb-0">{step.container}</p>
                                    </div>
                                </TabPanel>
                            </TabPanels>
                        </Tabs>
                    )}
                    {step.skipped && (
                        <p className="mb-0">
                            <strong>
                                Step has been skipped
                                {cachedResultFound && <> because a cached result was found for this workspace</>}.
                            </strong>
                        </p>
                    )}
                </div>
            </div>
        </Collapsible>
    )
}

interface StepStateIconProps {
    step: BatchSpecWorkspaceStepFields
}
const StepStateIcon: React.FunctionComponent<StepStateIconProps> = ({ step }) => {
    if (step.cachedResultFound) {
        return (
            <ContentSaveIcon
                className="icon-inline text-success"
                data-tooltip="A cached result for this step has been found"
            />
        )
    }
    if (step.skipped) {
        return <LinkVariantRemoveIcon className="icon-inline text-muted" data-tooltip="The step has been skipped" />
    }
    if (!step.startedAt) {
        return <TimerSandIcon className="icon-inline text-muted" data-tooltip="Waiting to be processed" />
    }
    if (!step.finishedAt) {
        return <LoadingSpinner className="icon-inline text-muted" data-tooltip="Currently running" />
    }
    if (step.exitCode === 0) {
        return <CheckCircleIcon className="icon-inline text-success" data-tooltip="Ran successfully" />
    }
    return (
        <AlertCircleIcon
            className="icon-inline text-danger"
            data-tooltip={`Step failed with exit code ${String(step.exitCode)}`}
        />
    )
}

const StepTimer: React.FunctionComponent<{ step: BatchSpecWorkspaceStepFields }> = ({ step }) => {
    if (!step.startedAt) {
        return null
    }
    return <Duration start={step.startedAt} end={step.finishedAt ?? undefined} />
}

const WorkspaceStepFileDiffConnection: React.FunctionComponent<
    {
        workspaceID: Scalars['ID']
        step: number
    } & ThemeProps
> = ({ workspaceID, step, isLightTheme }) => {
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryBatchSpecWorkspaceStepFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                node: workspaceID,
                step,
            }),
        [workspaceID, step]
    )
    const history = useHistory()
    return (
        <FileDiffConnection
            listClassName="list-group list-group-flush"
            noun="changed file"
            pluralNoun="changed files"
            queryConnection={queryFileDiffs}
            nodeComponent={FileDiffNode}
            nodeComponentProps={{
                history,
                location: history.location,
                isLightTheme,
                persistLines: true,
                lineNumbers: true,
            }}
            defaultFirst={15}
            hideSearch={true}
            noSummaryIfAllNodesVisible={true}
            history={history}
            location={history.location}
            useURLQuery={false}
            cursorPaging={true}
        />
    )
}
