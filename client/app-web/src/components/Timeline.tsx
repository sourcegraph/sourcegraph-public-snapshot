import classNames from 'classnames'
import { formatDistance } from 'date-fns/esm'
import React, { FunctionComponent, ReactNode } from 'react'
import { Collapsible } from './Collapsible'
import { Timestamp } from './time/Timestamp'

export interface TimelineStage {
    icon: ReactNode
    text: ReactNode
    details?: ReactNode
    date?: string | null
    className?: string
    expanded?: boolean
}

export interface TimelineProps {
    stages: TimelineStage[]
    now?: () => Date
    className?: string
}

export const Timeline: FunctionComponent<TimelineProps> = ({ stages, now, className }) => (
    <>
        <h3>Timeline</h3>

        <div className={className}>
            {stages.map((stage, stageIndex) => {
                if (!stage.date) {
                    return null
                }

                const previousDate = stages.map(stage => stage.date).find((date, index) => !!date && index < stageIndex)

                const meta = <TimelineMeta stage={{ ...stage, date: stage.date }} now={now} />

                return (
                    // Use index as key because the values in each step may not be unique. This is
                    // OK here because this list will not be updated during this component's lifetime.
                    /* eslint-disable react/no-array-index-key */
                    <div key={stageIndex}>
                        {previousDate && (
                            <div className="d-flex align-items-center">
                                <div className="flex-0">
                                    <div className="timeline__executor-task-separator" />
                                </div>
                                <div className="flex-1">
                                    <span className="text-muted ml-4">
                                        {formatDistance(new Date(stage.date), new Date(previousDate))}
                                    </span>
                                </div>
                            </div>
                        )}

                        {stage.details ? (
                            <>
                                <Collapsible
                                    title={meta}
                                    className="p-0 font-weight-normal"
                                    buttonClassName="mb-0"
                                    titleAtStart={true}
                                    defaultExpanded={stage.expanded}
                                >
                                    <div className="timeline__executor-task-details">{stage.details}</div>
                                </Collapsible>
                            </>
                        ) : (
                            meta
                        )}
                    </div>
                )
            })}
        </div>
    </>
)

export interface TimelineMetaProps {
    stage: TimelineStage & { date: string }
    now?: () => Date
}

export const TimelineMeta: FunctionComponent<TimelineMetaProps> = ({ stage, now }) => (
    <>
        <div className="d-flex align-items-center">
            <div className="flex-0 m-2">
                <div className={classNames('timeline__executor-task-icon', stage.className)}>{stage.icon}</div>
            </div>
            <div className="flex-1">
                {stage.text} <Timestamp date={stage.date} now={now} noAbout={true} />
            </div>
        </div>
    </>
)
