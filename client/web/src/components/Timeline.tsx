import classNames from 'classnames'
import { formatDistance } from 'date-fns/esm'
import React, { FunctionComponent, ReactNode } from 'react'
import { Timestamp } from './time/Timestamp'

export interface TimelineStage {
    icon: ReactNode
    text: ReactNode
    date?: string | null
    className?: string
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
                const previousDate = stages
                    .map(stage => stage.date)
                    .filter((date, index) => !!date && index < stageIndex)
                    .pop()

                return (
                    stage.date && (
                        <>
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
                        </>
                    )
                )
            })}
        </div>
    </>
)
