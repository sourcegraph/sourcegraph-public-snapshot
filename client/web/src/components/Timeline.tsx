import { FunctionComponent, ReactNode } from 'react'

import classNames from 'classnames'
import { formatDistance } from 'date-fns/esm'

import { Collapsible } from './Collapsible'
import { Timestamp } from './time/Timestamp'

import styles from './Timeline.module.scss'

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

export const Timeline: FunctionComponent<React.PropsWithChildren<TimelineProps>> = ({ stages, now, className }) => (
    <div className={className}>
        {stages.map((stage, stageIndex) => {
            if (!stage.date) {
                return null
            }

            const previousDate = stages
                .map(stage => stage.date)
                .filter((date, index) => !!date && index < stageIndex)
                .reverse()?.[0]

            const meta = <TimelineMeta stage={{ ...stage, date: stage.date }} now={now} />

            return (
                // Use index as key because the values in each step may not be unique. This is
                // OK here because this list will not be updated during this component's lifetime.
                /* eslint-disable react/no-array-index-key */
                <div key={stageIndex}>
                    {previousDate && (
                        <div className="d-flex align-items-center">
                            <div className="flex-0">
                                <div className={styles.executorTaskSeparator} />
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
                                <div className={styles.executorTaskDetails}>{stage.details}</div>
                            </Collapsible>
                        </>
                    ) : (
                        meta
                    )}
                </div>
            )
        })}
    </div>
)

export interface TimelineMetaProps {
    stage: TimelineStage & { date: string }
    now?: () => Date
}

export const TimelineMeta: FunctionComponent<React.PropsWithChildren<TimelineMetaProps>> = ({ stage, now }) => (
    <>
        <div className="d-flex align-items-center">
            <div className="flex-0 m-2">
                <div className={classNames(styles.executorTaskIcon, stage.className)}>{stage.icon}</div>
            </div>
            <div className="flex-1">
                {stage.text}{' '}
                <span className="text-muted">
                    <Timestamp date={stage.date} now={now} noAbout={true} />
                </span>
            </div>
        </div>
    </>
)
