import { type FunctionComponent, useMemo } from 'react'

import { mdiAlertCircle, mdiCheck, mdiCheckCircle, mdiFileUpload, mdiProgressClock, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { isDefined } from '@sourcegraph/common'
import { Card, CardBody, Code, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { Collapsible } from '../../../../components/Collapsible'
import { LogOutput } from '../../../../components/LogOutput'
import { Timeline, type TimelineStage } from '../../../../components/Timeline'
import { type PreciseIndexFields, PreciseIndexState } from '../../../../graphql-operations'
import { formatDurationLong } from '../../../../util/time'

import styles from './IndexTimeline.module.scss'

export interface IndexTimelineProps {
    index: PreciseIndexFields
    now?: () => Date
    className?: string
}

export const IndexTimeline: FunctionComponent<IndexTimelineProps> = ({ index, now, className }) => {
    const stages = useMemo(() => {
        const stages: TimelineStage[] = []

        // Stage: queued for indexing
        if (index.queuedAt) {
            stages.push({
                icon: <Icon aria-label="Success" svgPath={mdiTimerSand} />,
                text: 'Queued for indexing',
                date: index.queuedAt,
                className: 'bg-success',
            })
        }

        // Stage: indexing started
        if (index.indexingStartedAt) {
            stages.push({
                icon: <Icon aria-label="Success" svgPath={mdiProgressClock} />,
                text: 'Began indexing',
                date: index.indexingStartedAt,
                className: 'bg-success',
            })
        }

        // Stage: indexing job steps
        let stage = indexSetupStage(index, now)
        if (stage) {
            stages.push(stage)
        }
        stage = indexPreIndexStage(index, now)
        if (stage) {
            stages.push(stage)
        }
        stage = indexIndexStage(index, now)
        if (stage) {
            stages.push(stage)
        }
        stage = indexUploadStage(index, now)
        if (stage) {
            stages.push(stage)
        }
        stage = indexTeardownStage(index, now)
        if (stage) {
            stages.push(stage)
        }

        // Stage: Indexing failed (shown conditionally)
        //
        // Do not distinctly show the end of indexing unless it was a failure that produced
        // to submit an upload record. If we did submit a record, then the end result of this
        // job is successful to the user (if processing succeeds).
        if (index.indexingFinishedAt && index.state === PreciseIndexState.INDEXING_ERRORED) {
            stages.push({
                icon: <Icon aria-label="" svgPath={mdiAlertCircle} />,
                text: 'Failed indexing',
                date: index.indexingFinishedAt,
                className: 'bg-danger',
            })
        }

        // Stage: Manual upload, or indexing job uploaded artifact
        if (index.uploadedAt) {
            if (index.state === PreciseIndexState.UPLOADING_INDEX) {
                stages.push({
                    icon: <Icon aria-label="Success" svgPath={mdiFileUpload} />,
                    text: 'Began uploading',
                    date: index.uploadedAt,
                    className: 'bg-success',
                })
            } else if (index.state === PreciseIndexState.PROCESSING_ERRORED) {
                if (!index.processingStartedAt) {
                    stages.push({
                        icon: <Icon aria-label="" svgPath={mdiAlertCircle} />,
                        text: 'Uploading failed',
                        date: index.uploadedAt,
                        className: 'bg-danger',
                    })
                }
            } else {
                stages.push({
                    icon: <Icon aria-label="Success" svgPath={mdiTimerSand} />,
                    text: 'Queued for processing',
                    date: index.uploadedAt,
                    className: 'bg-success',
                })
            }
        }

        // Stage: Post-upload processing stated
        if (index.processingStartedAt) {
            stages.push({
                icon: <Icon aria-label="Success" svgPath={mdiProgressClock} />,
                text: 'Began processing',
                date: index.processingStartedAt,
                className: 'bg-success',
            })
        }

        // Stage: Processing terminated (success or failure)
        if (index.processingFinishedAt) {
            if (index.state === PreciseIndexState.PROCESSING_ERRORED) {
                if (index.processingStartedAt) {
                    stages.push({
                        icon: <Icon aria-label="Failed" svgPath={mdiAlertCircle} />,
                        text: 'Failed',
                        date: index.processingFinishedAt,
                        className: 'bg-danger',
                    })
                }
            } else {
                stages.push({
                    icon: <Icon aria-label="Success" svgPath={mdiCheck} />,
                    text: 'Finished',
                    date: index.processingFinishedAt,
                    className: 'bg-success',
                })
            }
        }

        return stages
    }, [index, now])

    return <Timeline stages={stages} now={now} className={className} />
}

const indexSetupStage = (index: PreciseIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps || index.steps.setup.length === 0
        ? undefined
        : {
              text: 'Setup',
              details: index.steps.setup.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(index.steps.setup),
          }

const indexPreIndexStage = (index: PreciseIndexFields, now?: () => Date): TimelineStage | undefined => {
    if (!index.steps) {
        return undefined
    }

    const logEntries = index.steps.preIndex.map(step => step.logEntry).filter(isDefined)

    return logEntries.length === 0
        ? undefined
        : {
              text: 'Pre Index',
              details: index.steps.preIndex.map(
                  step =>
                      step.logEntry && (
                          <div key={`${step.image}${step.root}${step.commands.join(' ')}}`}>
                              <ExecutionLogEntry logEntry={step.logEntry} now={now}>
                                  <ExecutionMetaInformation
                                      {...{
                                          image: step.image,
                                          commands: step.commands,
                                          root: step.root,
                                      }}
                                  />
                              </ExecutionLogEntry>
                          </div>
                      )
              ),
              ...genericStage(logEntries),
          }
}

const indexIndexStage = (index: PreciseIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps?.index?.logEntry
        ? undefined
        : {
              text: 'Index',
              details: (
                  <>
                      <ExecutionLogEntry logEntry={index.steps.index.logEntry} now={now}>
                          <ExecutionMetaInformation
                              {...{
                                  image: index.inputIndexer,
                                  commands: index.steps.index.indexerArgs,
                                  root: index.inputRoot,
                              }}
                          />
                      </ExecutionLogEntry>
                  </>
              ),
              ...genericStage(index.steps.index.logEntry),
          }

const indexUploadStage = (index: PreciseIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index?.steps?.upload
        ? undefined
        : {
              text: 'Upload',
              details: <ExecutionLogEntry logEntry={index.steps.upload} now={now} />,
              ...genericStage(index.steps.upload),
          }

const indexTeardownStage = (index: PreciseIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps || index.steps.teardown.length === 0
        ? undefined
        : {
              text: 'Teardown',
              details: index.steps.teardown.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(index.steps.teardown),
          }

const genericStage = <E extends { startTime: string; exitCode: number | null }>(
    value: E | E[]
): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expandedByDefault'> => {
    const finished = Array.isArray(value)
        ? value.every(logEntry => logEntry.exitCode !== null)
        : value.exitCode !== null
    const success = Array.isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: !finished ? (
            <Icon aria-label="Success" svgPath={mdiProgressClock} />
        ) : success ? (
            <Icon aria-label="Success" svgPath={mdiCheck} />
        ) : (
            <Icon aria-label="Failed" svgPath={mdiAlertCircle} />
        ),
        date: Array.isArray(value) ? value[0].startTime : value.startTime,
        className: success || !finished ? 'bg-success' : 'bg-danger',
        expandedByDefault: !(success || !finished),
    }
}

interface ExecutionLogEntryProps {
    logEntry: {
        key: string
        command: string[]
        startTime: string
        exitCode: number | null
        out: string
        durationMilliseconds: number | null
    }
    now?: () => Date
}

const ExecutionLogEntry: React.FunctionComponent<React.PropsWithChildren<ExecutionLogEntryProps>> = ({
    logEntry,
    children,
    now,
}) => (
    <Card className="mb-3">
        <CardBody>
            {logEntry.command.length > 0 ? (
                <LogOutput text={logEntry.command.join(' ')} className="mb-3" logDescription="Executed command:" />
            ) : (
                <div className="mb-3">
                    <span className="text-muted">Internal step {logEntry.key}.</span>
                </div>
            )}

            <div>
                {logEntry.exitCode === null && <LoadingSpinner className="mr-1" />}
                {logEntry.exitCode !== null && (
                    <>
                        {logEntry.exitCode === 0 ? (
                            <Icon
                                className="text-success mr-1"
                                svgPath={mdiCheckCircle}
                                inline={false}
                                aria-label="Success"
                            />
                        ) : (
                            <Icon
                                className="text-danger mr-1"
                                svgPath={mdiAlertCircle}
                                inline={false}
                                aria-label="Failed"
                            />
                        )}
                    </>
                )}
                <span className="text-muted">Started</span>{' '}
                <Timestamp date={logEntry.startTime} now={now} noAbout={true} />
                {logEntry.exitCode !== null && logEntry.durationMilliseconds !== null && (
                    <>
                        <span className="text-muted">, ran for</span>{' '}
                        {formatDurationLong(logEntry.durationMilliseconds)}
                    </>
                )}
            </div>
            {children}
        </CardBody>

        <div className="p-2">
            {logEntry.out ? (
                <Collapsible
                    title="Log output"
                    titleAtStart={true}
                    className="p-0"
                    buttonClassName={styles.collapseButton}
                    defaultExpanded={logEntry.exitCode !== null && logEntry.exitCode !== 0}
                >
                    <LogOutput text={logEntry.out} logDescription="Log output:" />
                </Collapsible>
            ) : (
                <div className="p-2">
                    <span className="text-muted">No log output available.</span>
                </div>
            )}
        </div>
    </Card>
)

interface ExecutionMetaInformationProps {
    image: string
    commands: string[]
    root: string
}

const ExecutionMetaInformation: React.FunctionComponent<ExecutionMetaInformationProps> = ({
    image,
    commands,
    root,
}) => (
    <div className="pt-3">
        <div className={classNames(styles.dockerCommandSpec, 'py-2 border-top pl-2')}>
            <strong className={styles.header}>Image</strong>
            <div>{image}</div>
        </div>
        <div className={classNames(styles.dockerCommandSpec, 'py-2 border-top pl-2')}>
            <strong className={styles.header}>Commands</strong>
            <div>
                <Code>{commands.join(' ')}</Code>
            </div>
        </div>
        <div className={classNames(styles.dockerCommandSpec, 'py-2 border-top pl-2')}>
            <strong className={styles.header}>Root</strong>
            <div>/{root}</div>
        </div>
    </div>
)
