import React, { type FunctionComponent, type ReactNode, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import VisuallyHidden from '@reach/visually-hidden'
import classNames from 'classnames'
import { formatDistance } from 'date-fns'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Button, Collapse, CollapseHeader, CollapsePanel, Icon } from '@sourcegraph/wildcard'

import styles from './Timeline.module.scss'

export interface TimelineStage {
    icon: ReactNode
    text: ReactNode
    details?: ReactNode
    date: string
    className?: string
    expandedByDefault?: boolean
}

export interface TimelineProps {
    stages: TimelineStage[]
    className?: string
    showDurations?: boolean
    /** For testing only */
    now?: () => Date
}

export const Timeline: FunctionComponent<React.PropsWithChildren<TimelineProps>> = ({
    stages,
    now,
    className,
    showDurations = true,
}) => (
    <div className={classNames('w-100', className)}>
        {stages.map((stage, stageIndex) => (
            <span key={stageIndex}>
                {stageIndex !== 0 && (
                    <div className="d-flex align-items-center">
                        <div className={styles.separator} />
                        {showDurations && (
                            <span className="flex-1 text-muted ml-4">
                                <VisuallyHidden>Step took</VisuallyHidden>
                                {formatDistance(new Date(stage.date), new Date(stages[stageIndex - 1]?.date))}
                            </span>
                        )}
                    </div>
                )}
                <TimelineStage key={`${stage.text}+${stage.date}`} stage={stage} now={now} />
            </span>
        ))}
    </div>
)

interface TimelineStageProps {
    stage: TimelineStage
    now?: () => Date
}

const TimelineStage: FunctionComponent<React.PropsWithChildren<TimelineStageProps>> = ({
    stage: { className, details, date, expandedByDefault = false, icon, text },
    now,
}) => {
    const [isExpanded, setIsExpanded] = useState(expandedByDefault)

    const stageLabel = (
        <div className="d-flex align-items-center">
            <div className={classNames(styles.icon, className)}>{icon}</div>
            <div className="flex-1">
                {text}
                <span className="text-muted ml-1">
                    <Timestamp date={date} now={now} noAbout={true} />
                </span>
            </div>
        </div>
    )

    return details ? (
        <Collapse isOpen={isExpanded} onOpenChange={setIsExpanded} openByDefault={expandedByDefault}>
            <CollapseHeader
                as={Button}
                className="p-0 m-0 border-0 w-100 font-weight-normal d-flex justify-content-between align-items-center"
            >
                {stageLabel}
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} className="mr-1" />
            </CollapseHeader>
            <CollapsePanel className={styles.details}>{details}</CollapsePanel>
        </Collapse>
    ) : (
        stageLabel
    )
}
