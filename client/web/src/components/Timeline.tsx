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
            {stages
                .map((stage, stageIndex) => ({ stage, stageIndex }))
                .map(({ stage, stageIndex }) => {
                    if (!stage.date) {
                        return null
                    }

                    const previousDate = stages
                        .map(stage => stage.date)
                        .filter((date, index) => !!date && index < stageIndex)
                        .pop()

                    const meta = (
                        <div className="d-flex align-items-center">
                            <div className="flex-0 m-2">
                                <div className={classNames('timeline__executor-task-icon', stage.className)}>
                                    {stage.icon}
                                </div>
                            </div>
                            <div className="flex-1">
                                {stage.text} <Timestamp date={stage.date} now={now} noAbout={true} />
                            </div>
                        </div>
                    )

                    return (
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
