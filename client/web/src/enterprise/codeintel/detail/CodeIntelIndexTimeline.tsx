import { isArray } from 'lodash'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import classNames from 'classnames'
import ErrorIcon from 'mdi-react/ErrorIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React, { FunctionComponent, useMemo } from 'react'
import { isDefined } from '../../../../../shared/src/util/types'
import { Collapsible } from '../../../components/Collapsible'
import { Timestamp } from '../../../components/time/Timestamp'
import { Timeline, TimelineStage } from '../../../components/Timeline'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'

export interface CodeIntelIndexTimelineProps {
    index: LsifIndexFields
    now?: () => Date
    className?: string
}

export const CodeIntelIndexTimeline: FunctionComponent<CodeIntelIndexTimelineProps> = ({ index, now, className }) => {
    const stages = useMemo(
        () => [
            { icon: <TimerSandIcon />, text: 'Queued', date: index.queuedAt, className: 'bg-success' },
            { icon: <ProgressClockIcon />, text: 'Began processing', date: index.startedAt, className: 'bg-success' },

            indexSetupStage(index, now),
            indexPreIndexStage(index, now),
            indexIndexStage(index, now),
            indexUploadStage(index, now),
            indexTeardownStage(index, now),

            index.state === LSIFIndexState.COMPLETED
                ? { icon: <CheckIcon />, text: 'Finished', date: index.finishedAt, className: 'bg-success' }
                : { icon: <ErrorIcon />, text: 'Failed', date: index.finishedAt, className: 'bg-danger' },
        ],
        [index, now]
    )

    return <Timeline stages={stages.filter(isDefined)} now={now} className={className} />
}

const indexSetupStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    index.steps.setup.length === 0
        ? undefined
        : {
              text: 'Setup',
              details: index.steps.setup.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(index.steps.setup),
          }

const indexPreIndexStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined => {
    const logEntries = index.steps.preIndex.map(step => step.logEntry).filter(isDefined)

    return logEntries.length === 0
        ? undefined
        : {
              text: 'Pre Index',
              details: index.steps.preIndex.map(
                  step =>
                      step.logEntry && (
                          <div key={`${step.image}${step.root}${step.commands.join(' ')}}`}>
                              <ExecutionLogEntry
                                  logEntry={step.logEntry}
                                  now={now}
                                  meta={{
                                      image: step.image,
                                      commands: step.commands,
                                      root: step.root,
                                  }}
                              />
                          </div>
                      )
              ),
              ...genericStage(logEntries),
          }
}

const indexIndexStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps.index.logEntry
        ? undefined
        : {
              text: 'Index',
              details: (
                  <>
                      <ExecutionLogEntry
                          logEntry={index.steps.index.logEntry}
                          now={now}
                          meta={{
                              image: index.inputIndexer,
                              commands: index.steps.index.indexerArgs,
                              root: index.inputRoot,
                          }}
                      />
                  </>
              ),
              ...genericStage(index.steps.index.logEntry),
          }

const indexUploadStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    !index.steps.upload
        ? undefined
        : {
              text: 'Upload',
              details: <ExecutionLogEntry logEntry={index.steps.upload} now={now} />,
              ...genericStage(index.steps.upload),
          }

const indexTeardownStage = (index: LsifIndexFields, now?: () => Date): TimelineStage | undefined =>
    index.steps.teardown.length === 0
        ? undefined
        : {
              text: 'Teardown',
              details: index.steps.teardown.map(logEntry => (
                  <ExecutionLogEntry key={logEntry.key} logEntry={logEntry} now={now} />
              )),
              ...genericStage(index.steps.teardown),
          }

const genericStage = <E extends { startTime: string; exitCode: number }>(
    value: E | E[]
): Pick<TimelineStage, 'icon' | 'date' | 'className' | 'expanded'> => {
    const success = isArray(value) ? value.every(logEntry => logEntry.exitCode === 0) : value.exitCode === 0

    return {
        icon: success ? <CheckIcon /> : <ErrorIcon />,
        date: isArray(value) ? value[0].startTime : value.startTime,
        className: success ? 'bg-success' : 'bg-danger',
        expanded: !success,
    }
}

interface ExecutionLogEntryProps {
    logEntry: {
        key: string
        command: string[]
        startTime: string
        exitCode: number
        out: string
        durationMilliseconds: number
    }
    meta?: {
        image: string
        commands: string[]
        root: string
    }
    now?: () => Date
}

const ExecutionLogEntry: FunctionComponent<ExecutionLogEntryProps> = ({ logEntry, meta, now }) => (
    <div className="card mb-3">
        <div className="card-body">
            <LogOutput text={logEntry.command.join(' ')} className="mb-3" />

            <div>
                {logEntry.exitCode === 0 ? (
                    <CheckCircleIcon className="text-success" />
                ) : (
                    <ErrorIcon className="text-danger" />
                )}

                <span className="ml-2">
                    <span className="text-muted">Started</span>{' '}
                    <Timestamp date={logEntry.startTime} now={now} noAbout={true} />
                    <span className="text-muted">, ran for</span> {formatMilliseconds(logEntry.durationMilliseconds)}
                </span>
            </div>

            {meta && (
                <table className="table mt-4 mb-0 docker-command-spec">
                    <tr>
                        <th>Image</th>
                        <td>{meta.image}</td>
                    </tr>
                    <tr>
                        <th>Commands</th>
                        <td>
                            <code>{meta.commands.join(' ')}</code>
                        </td>
                    </tr>
                    <tr>
                        <th>Root</th>
                        <td>/{meta.root}</td>
                    </tr>
                </table>
            )}
        </div>

        <div className="p-2">
            {logEntry.out ? (
                <Collapsible title="Log output" titleAtStart={true} buttonClassName="p-2">
                    <LogOutput text={logEntry.out} />
                </Collapsible>
            ) : (
                <div className="p-2">
                    <span className="text-muted">No log output available.</span>
                </div>
            )}
        </div>
    </div>
)

interface LogOutputProps {
    text: string
    className?: string
}

const LogOutput: FunctionComponent<LogOutputProps> = ({ text, className }) => (
    <pre className={classNames('bg-code rounded p-3 mb-0', className)}>
        <code>
            {/* {text.split('\n').map(line => line.replace(/^std(out|err): /, '')).join('\n')} */}
            {text}
        </code>
    </pre>
)

const formatMilliseconds = (milliseconds: number): string => {
    if (milliseconds < 1000) {
        return `${milliseconds} milliseconds`
    }
    if (milliseconds < 1000 * 60) {
        return `${milliseconds / 1000} seconds`
    }
    if (milliseconds < 1000 * 60 * 60) {
        return `${milliseconds / (1000 * 60)} minutes`
    }

    return `${milliseconds / (1000 * 60 * 60)} hours`
}
