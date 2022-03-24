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

const Header: React.FunctionComponent<{ onClose: () => void }> = ({ children, onClose }) => (
    <div className="d-flex justify-content-between align-items-start">
        <p className={styles.title}>Quick start</p>
        <Button variant="icon" data-testid="tour-close-btn" onClick={onClose}>
            <Icon as={CloseIcon} /> {children}
        </Button>
    </div>
)

export const TourContent: React.FunctionComponent<TourContentProps> = ({
    onClose,
    tasks,
    variant,
    className,
    height = 18,
}) => {
    const { completedCount, totalCount, completedTaskChunks, completedTasks, ongoingTasks } = useMemo(() => {
        const completedTasks = tasks.filter(task => task.completed === 100)
        return {
            completedTasks,
            ongoingTasks: tasks.filter(task => task.completed !== 100),
            completedTaskChunks: chunk(completedTasks, 3),
            totalCount: tasks.filter(task => typeof task.completed === 'number').length,
            completedCount: completedTasks.length,
        }
    }, [tasks])

    return (
        <div className={className} data-testid="tour-content">
            {variant === 'horizontal' && <Header onClose={onClose}>Don't show again</Header>}
            <MarketingBlock
                wrapperClassName={classNames('w-100 d-flex', variant !== 'horizontal' && styles.marketingBlockWrapper)}
                contentClassName={classNames(styles.marketingBlockContent, 'w-100 py-3 d-flex flex-column')}
            >
                {variant !== 'horizontal' && <Header onClose={onClose} />}
                <div
                    className={classNames(
                        styles.taskList,
                        variant && styles[`is${upperFirst(variant)}` as keyof typeof styles]
                    )}
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ maxHeight: `${height}rem` }}
                >
                    {variant === 'horizontal' &&
                        completedTaskChunks.map((completedTaskChunk, index) => (
                            <div key={index}>
                                {variant === 'horizontal' && index === 0 && <p className={styles.title}>Completed</p>}
                                {completedTaskChunk.map(completedTask => (
                                    <div key={completedTask.title}>
                                        <Icon as={CheckCircleIcon} size="sm" className="text-success" />
                                        <span className="ml-1">{completedTask.title}</span>
                                    </div>
                                ))}
                            </div>
                        ))}
                    {ongoingTasks.map(task => (
                        <TourTask key={task.title} {...task} variant={variant !== 'horizontal' ? 'small' : undefined} />
                    ))}
                    {variant !== 'horizontal' && (
                        <div>
                            {completedTasks.map(completedTask => (
                                <div key={completedTask.title}>
                                    <Icon as={CheckCircleIcon} size="sm" className="text-success" />
                                    <span className="ml-1">{completedTask.title}</span>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </MarketingBlock>
            <p className="text-right mt-2 mb-0">
                <Icon as={CheckCircleIcon} className="text-success" />
                {completedCount} of {totalCount} completed
            </p>
        </div>
    )
}
