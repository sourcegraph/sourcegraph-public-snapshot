import React, { useMemo } from 'react'

import classNames from 'classnames'
import { chunk, upperFirst } from 'lodash'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../../components/MarketingBlock'

import { TourTask } from './TourTask'
import { TourTaskType } from './types'

import styles from './Tour.module.scss'

interface TourContentProps {
    tasks: (TourTaskType | TourTaskType)[]
    onClose: () => void
    variant?: 'horizontal'
    height?: number
    className?: string
}

const Header: React.FunctionComponent<React.PropsWithChildren<{ onClose: () => void }>> = ({ children, onClose }) => (
    <div className="d-flex justify-content-between align-items-start">
        <p className={styles.title}>Quick start</p>
        <Button variant="icon" data-testid="tour-close-btn" onClick={onClose}>
            <Icon as={CloseIcon} /> {children}
        </Button>
    </div>
)

const Footer: React.FunctionComponent<React.PropsWithChildren<{ completedCount: number; totalCount: number }>> = ({
    completedCount,
    totalCount,
}) => (
    <p className="text-right mt-2 mb-0">
        <Icon
            as={CheckCircleIcon}
            className={classNames('mr-1', completedCount === 0 ? 'text-muted' : 'text-success')}
        />
        {completedCount} of {totalCount} completed
    </p>
)

const CompletedItem: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <li className="d-flex align-items-start">
        <Icon as={CheckCircleIcon} size="sm" className={classNames('text-success mr-1', styles.completedCheckIcon)} />
        <span className="flex-1">{children}</span>
    </li>
)

export const TourContent: React.FunctionComponent<React.PropsWithChildren<TourContentProps>> = ({
    onClose,
    tasks,
    variant,
    className,
    height = 18,
}) => {
    const { completedCount, totalCount, completedTasks, completedTaskChunks, ongoingTasks } = useMemo(() => {
        const completedTasks = tasks.filter(task => task.completed === 100)
        return {
            completedTasks,
            ongoingTasks: tasks.filter(task => task.completed !== 100),
            completedTaskChunks: chunk(completedTasks, 3),
            totalCount: tasks.filter(task => typeof task.completed === 'number').length,
            completedCount: completedTasks.length,
        }
    }, [tasks])
    const isHorizontal = variant === 'horizontal'

    return (
        <div className={className} data-testid="tour-content">
            {isHorizontal && <Header onClose={onClose}>Don't show again</Header>}
            <MarketingBlock
                wrapperClassName={classNames('w-100 d-flex', !isHorizontal && styles.marketingBlockWrapper)}
                contentClassName={classNames(styles.marketingBlockContent, 'w-100 d-flex flex-column pt-3 pb-1')}
            >
                {!isHorizontal && <Header onClose={onClose} />}
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
                            <p className={styles.title}>Completed</p>
                            <div className={styles.completedItemsInner}>
                                {completedTaskChunks.map((completedTaskChunk, index) => (
                                    <ul key={index} className="p-0 m-0 list-unstyled text-nowrap">
                                        {completedTaskChunk.map(completedTask => (
                                            <CompletedItem key={completedTask.title}>
                                                {completedTask.title}
                                            </CompletedItem>
                                        ))}
                                    </ul>
                                ))}
                            </div>
                        </div>
                    )}
                    {ongoingTasks.map(task => (
                        <TourTask key={task.title} {...task} variant={!isHorizontal ? 'small' : undefined} />
                    ))}
                    {!isHorizontal && completedTasks.length > 0 && (
                        <div>
                            {completedTasks.map(completedTask => (
                                <CompletedItem key={completedTask.title}>{completedTask.title}</CompletedItem>
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
