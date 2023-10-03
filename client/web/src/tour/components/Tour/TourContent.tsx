import React, { useMemo } from 'react'

import { mdiClose, mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'
import { chunk, upperFirst } from 'lodash'

import type { TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import { Badge, Button, Icon, Text } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../../components/MarketingBlock'

import { TourTask } from './TourTask'

import styles from './Tour.module.scss'

interface TourContentProps {
    title?: string
    keepCompletedTasks?: boolean
    tasks: TourTaskType[]
    onClose?: () => void
    variant?: 'horizontal'
    height?: number
    className?: string
}

const Header: React.FunctionComponent<React.PropsWithChildren<{ onClose: () => void; title?: string }>> = ({
    children,
    onClose,
    title = 'Quick start',
}) => (
    <div className="d-flex align-items-start">
        <Text className={styles.title}>{title}</Text>
        <Badge className="ml-2" variant="warning">
            Experimental
        </Badge>
        <Button
            className="ml-auto"
            variant="icon"
            data-testid="tour-close-btn"
            onClick={onClose}
            aria-label="Close quick start"
        >
            <Icon aria-hidden={true} svgPath={mdiClose} /> {children}
        </Button>
    </div>
)

const Footer: React.FunctionComponent<React.PropsWithChildren<{ completedCount: number; totalCount: number }>> = ({
    completedCount,
    totalCount,
}) => (
    <Text alignment="right" className="mt-2 mb-0">
        <Icon
            className={classNames('mr-1', completedCount === 0 ? 'text-muted' : 'text-success')}
            aria-hidden={true}
            svgPath={mdiCheckCircle}
        />
        {completedCount} of {totalCount} completed
    </Text>
)

const CompletedItem: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <li className="d-flex align-items-start">
        <Icon
            size="sm"
            className={classNames('text-success mr-1', styles.completedCheckIcon)}
            aria-hidden={true}
            svgPath={mdiCheckCircle}
        />
        <span className="flex-1">{children}</span>
    </li>
)

export const TourContent: React.FunctionComponent<React.PropsWithChildren<TourContentProps>> = ({
    onClose,
    tasks,
    variant,
    className,
    title,
    keepCompletedTasks,
    height = 18,
}) => {
    const { completedCount, totalCount, completedTasks, completedTaskChunks, ongoingTasks } = useMemo(() => {
        const completedTasks = tasks.filter(task => task.completed === 100)
        if (keepCompletedTasks) {
            return {
                completedTasks: [],
                ongoingTasks: tasks,
                completedTaskChunks: [],
                totalCount: tasks.filter(task => typeof task.completed === 'number').length,
                completedCount: completedTasks.length,
            }
        }
        return {
            completedTasks,
            ongoingTasks: tasks.filter(task => task.completed !== 100),
            completedTaskChunks: chunk(completedTasks, 3),
            totalCount: tasks.filter(task => typeof task.completed === 'number').length,
            completedCount: completedTasks.length,
        }
    }, [keepCompletedTasks, tasks])
    const isHorizontal = variant === 'horizontal'

    return (
        <div className={className} data-testid="tour-content">
            {isHorizontal && onClose && (
                <Header onClose={onClose} title={title}>
                    Don't show again
                </Header>
            )}
            <MarketingBlock
                wrapperClassName={classNames('w-100 d-flex', !isHorizontal && styles.marketingBlockWrapper)}
                contentClassName={classNames(styles.marketingBlockContent, 'w-100 d-flex flex-column pt-3 pb-1')}
            >
                {!isHorizontal && onClose && <Header onClose={onClose} title={title} />}
                <div
                    className={classNames(
                        styles.taskList,
                        variant && styles[`is${upperFirst(variant)}` as keyof typeof styles]
                    )}
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ maxHeight: `${height}rem` }}
                >
                    {isHorizontal && completedTaskChunks.length > 0 && (
                        <div className={classNames('pl-2 flex-grow-1', styles.completedItems)}>
                            <Text className={styles.title}>Completed</Text>
                            <div className={styles.completedItemsInner}>
                                {completedTaskChunks.map((completedTaskChunk, index) => (
                                    <ul key={index} className="p-0 m-0 list-unstyled text-nowrap">
                                        {completedTaskChunk.map((completedTask, index) => (
                                            <CompletedItem key={`${completedTask.title}-${index}`}>
                                                {completedTask.title}
                                            </CompletedItem>
                                        ))}
                                    </ul>
                                ))}
                            </div>
                        </div>
                    )}
                    {ongoingTasks.map((task, index) => (
                        <TourTask
                            key={`${task.title}-${index}`}
                            {...task}
                            variant={!isHorizontal ? 'small' : undefined}
                        />
                    ))}
                    {!isHorizontal && completedTasks.length > 0 && (
                        <div>
                            {completedTasks.map((completedTask, index) => (
                                <CompletedItem key={`${completedTask.title}-${index}`}>
                                    {completedTask.title}
                                </CompletedItem>
                            ))}
                        </div>
                    )}
                </div>
                {!isHorizontal && <Footer completedCount={completedCount} totalCount={totalCount} />}
            </MarketingBlock>
            {isHorizontal && <Footer completedCount={completedCount} totalCount={totalCount} />}
        </div>
    )
}
