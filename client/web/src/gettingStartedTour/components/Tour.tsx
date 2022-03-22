import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { chunk, uniq, upperFirst } from 'lodash'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import { CircularProgressbar } from 'react-circular-progressbar'
import ReactDOM from 'react-dom'
import { useHistory, useLocation } from 'react-router-dom'

import { ModalVideo } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../components/MarketingBlock'
import { useQuickStartTourListState } from '../../stores/quickStartTourState'
import { GETTING_STARTED_TOUR_MARKER } from '../GettingStartedTourInfo'
import { buildURIMarkers, isExternalURL, parseURIMarkers } from '../utils'

import { TourContext } from './context'
import { LinkOrAnchor } from './LinkOrAnchor'
import { TourTaskType, TourLanguage, TourTaskStepType } from './types'

import styles from './Tour.module.scss'

interface LanguagePickerProps {
    onClose: () => void
    onSelect: (language: TourLanguage) => void
}

const LanguagePicker: React.FunctionComponent<LanguagePickerProps> = ({ onClose, onSelect }) => (
    <div>
        <div className="d-flex justify-content-between">
            <p className="mt-2">Please select a language:</p>
            <CloseIcon onClick={onClose} size="1rem" />
        </div>
        <div className="d-flex flex-wrap">
            {Object.values(TourLanguage).map(language => (
                <Button
                    key={language}
                    className={classNames('mr-1 my-1', styles.language)}
                    onClick={() => onSelect(language)}
                    size="sm"
                >
                    {language}
                </Button>
            ))}
        </div>
    </div>
)

const isLanguageRequired = (step: TourTaskStepType): boolean => typeof step.action.value !== 'string'
const getActionValue = (step: TourTaskStepType, language?: TourLanguage): string =>
    typeof step.action.value === 'string'
        ? buildURIMarkers(step.action.value, step.id)
        : language
        ? buildURIMarkers(step.action.value[language], step.id)
        : '#'

type TourTaskProps = TourTaskType & {
    variant?: 'small'
}

const TourTask: React.FunctionComponent<TourTaskProps> = ({ title, steps, completed, icon, variant }) => {
    const [selectedStep, setSelectedStep] = useState<TourTaskStepType>()
    const [showLanguagePicker, setShowLanguagePicker] = useState(false)
    const { language, onLanguageSelect, onStepClick, onRestart } = useContext(TourContext)

    const handleLinkClick = useCallback(
        (event: React.MouseEvent<HTMLAnchorElement>, step: TourTaskStepType) => {
            onStepClick(step, language)
            if (isLanguageRequired(step) && !language) {
                event.preventDefault()
                setShowLanguagePicker(true)
                setSelectedStep(step)
            }
        },
        [language, onStepClick]
    )

    const handleVideoToggle = useCallback(
        (isOpen: boolean, step: TourTaskStepType) => {
            if (isOpen) {
                onStepClick(step, language)
            }
        },
        [language, onStepClick]
    )

    const onLanguageClose = useCallback(() => setShowLanguagePicker(false), [])

    const history = useHistory()
    const handleLanguageSelect = useCallback(
        (language: TourLanguage) => {
            onLanguageSelect(language)
            setShowLanguagePicker(false)
            if (!selectedStep) {
                return
            }
            onStepClick(selectedStep, language)
            const url = getActionValue(selectedStep, language)
            if (isExternalURL(url)) {
                window.open(url, '_blank')
            } else {
                history.push(url)
            }
        },
        [onStepClick, onLanguageSelect, selectedStep, history]
    )

    if (showLanguagePicker) {
        return <LanguagePicker onClose={onLanguageClose} onSelect={handleLanguageSelect} />
    }

    const isMultiStep = steps.length > 1
    return (
        <div className={classNames(icon && [styles.task, variant === 'small' && styles.isSmall])}>
            {icon && variant !== 'small' && <span className={styles.taskIcon}>{icon}</span>}
            <div className="flex-grow-1">
                <div className="d-flex justify-content-between position-relative">
                    {icon && variant === 'small' && <span className={classNames(styles.taskIcon)}>{icon}</span>}
                    <p className={styles.title}>{title}</p>
                    {completed === 100 && <CheckCircleIcon className="icon-inline text-success" size="1rem" />}
                    {typeof completed === 'number' && completed < 100 && (
                        <CircularProgressbar className={styles.progressBar} strokeWidth={10} value={completed || 0} />
                    )}
                </div>
                <ul className={classNames(styles.stepList, 'm-0', isMultiStep && styles.isMultiStep)}>
                    {steps.map(step => (
                        <li key={step.id} className={classNames(styles.stepListItem, 'd-flex align-items-center')}>
                            {step.action.type === 'link' && (
                                <LinkOrAnchor
                                    className="flex-grow-1"
                                    href={getActionValue(step, language)}
                                    onClick={event => handleLinkClick(event, step)}
                                >
                                    {step.label}
                                </LinkOrAnchor>
                            )}
                            {step.action.type === 'restart' && (
                                <div className="flex-grow">
                                    <p className="m-0">{step.label}</p>
                                    <div className="d-flex flex-column">
                                        <Button
                                            variant="link"
                                            className="align-self-start text-left pl-0"
                                            onClick={() => onRestart(step)}
                                        >
                                            {step.action.value}
                                        </Button>
                                    </div>
                                </div>
                            )}
                            {step.action.type === 'video' && (
                                <ModalVideo
                                    id={step.id}
                                    showCaption={true}
                                    title={step.label}
                                    className="flex-grow-1"
                                    titleClassName="text-left p-0 m-0"
                                    src={getActionValue(step, language)}
                                    onToggle={isOpen => handleVideoToggle(isOpen, step)}
                                />
                            )}
                            {isMultiStep && step.isCompleted && (
                                <CheckCircleIcon className={classNames('icon-inline', 'text-success')} size="1rem" />
                            )}
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    )
}

interface TourContentProps {
    tasks: (TourTaskType | TourTaskType)[]
    onClose: () => void
    variant?: 'horizontal'
    height?: number
    className?: string
}

const TourContentHeader: React.FunctionComponent<{ onClose: () => void }> = ({ children, onClose }) => (
    <div className="d-flex justify-content-between">
        <p className={styles.title}>Quick start</p>
        <span className="cursor-pointer" role="button" data-testid="tour-close-btn" onClick={onClose}>
            <CloseIcon size="1rem" /> {children}
        </span>
    </div>
)

const TourContent: React.FunctionComponent<TourContentProps> = ({
    onClose,
    tasks,
    variant,
    className,
    height = 18,
}) => {
    const { completedTaskChunks, completedTasks, ongoingTasks } = useMemo(() => {
        const completedTasks = tasks.filter(task => task.completed === 100)
        return {
            completedTaskChunks: chunk(completedTasks, 3),
            completedTasks,
            ongoingTasks: tasks.filter(task => task.completed !== 100),
        }
    }, [tasks])

    return (
        <div className={className} data-testid="tour-content">
            {variant === 'horizontal' && <TourContentHeader onClose={onClose}>Don't show again</TourContentHeader>}
            <MarketingBlock
                wrapperClassName={classNames('w-100 d-flex', variant !== 'horizontal' && styles.marketingBlockWrapper)}
                contentClassName={classNames(styles.marketingBlockContent, 'w-100 py-3 d-flex flex-column')}
            >
                {variant !== 'horizontal' && <TourContentHeader onClose={onClose} />}
                <div
                    className={classNames(styles.taskList, variant && styles[`is${upperFirst(variant)}`])}
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ maxHeight: `${height}rem` }}
                >
                    {variant === 'horizontal' &&
                        completedTaskChunks.map((completedTaskChunk, index) => (
                            <div key={index}>
                                {variant === 'horizontal' && index === 0 && <p className={styles.title}>Completed</p>}
                                {completedTaskChunk.map(completedTask => (
                                    <div key={completedTask.title}>
                                        <CheckCircleIcon className="icon-inline text-success" size="1rem" />
                                        <span className="ml-1">{completedTask.title}</span>
                                    </div>
                                ))}
                            </div>
                        ))}
                    {ongoingTasks.map(task => (
                        <TourTask key={task.title} {...task} variant={variant !== 'horizontal' ? 'small' : undefined} />
                    ))}
                    {variant !== 'horizontal' &&
                        completedTasks.map(completedTask => (
                            <div key={completedTask.title}>
                                <CheckCircleIcon className="icon-inline text-success" size="1rem" />
                                <span className="ml-1">{completedTask.title}</span>
                            </div>
                        ))}
                </div>
            </MarketingBlock>
            <p className="text-right mt-2 mb-0">
                <CheckCircleIcon className="icon-inline text-success" size="1rem" /> {completedTasks.length} of{' '}
                {tasks.length} completed
            </p>
        </div>
    )
}

interface TourAgentProps extends TelemetryProps {
    tasks: TourTaskType[]
    onStepComplete: (step: TourTaskStepType) => void
}
export const TourAgent: React.FunctionComponent<TourAgentProps> = React.memo(
    ({ tasks, telemetryService, onStepComplete }) => {
        // Agent 1: Track completion
        useEffect(() => {
            const filteredSteps = tasks.flatMap(task => task.steps).filter(step => step.completeAfterEvents)
            telemetryService?.addEventLogListener?.(eventName => {
                const step = filteredSteps.find(step => step.completeAfterEvents?.includes(eventName))
                if (step) {
                    onStepComplete(step)
                }
            })
        }, [telemetryService, tasks, onStepComplete])

        // Agent 2: Track info panel
        const [info, setInfo] = useState<TourTaskStepType['info'] | undefined>()

        const location = useLocation()

        useEffect(() => {
            const { isTour, stepId } = parseURIMarkers(location.search)
            if (!isTour || !stepId) {
                return
            }

            const info = tasks.flatMap(task => task.steps).find(step => stepId === step.id)?.info
            if (info) {
                setInfo(info)
            }
        }, [tasks, location])

        if (!info) {
            return null
        }

        const domNode = document.querySelector('.' + GETTING_STARTED_TOUR_MARKER)
        if (!domNode) {
            return null
        }

        return ReactDOM.createPortal(
            <div className={styles.infoPanel}>
                <CheckCircleIcon className={classNames('icon-inline', styles.infoIcon)} size="1rem" />
                <span dangerouslySetInnerHTML={{ __html: info }} />
            </div>,
            domNode
        )
    }
)

export type TourProps = TelemetryProps & {
    id: string
    tasks: TourTaskType[]
    extraTask?: TourTaskType
} & Pick<TourContentProps, 'variant' | 'className' | 'height'>

export const Tour: React.FunctionComponent<TourProps> = ({
    id: tourId,
    tasks,
    extraTask,
    telemetryService,
    ...props
}) => {
    const {
        completedStepIds = [],
        language,
        status,
        setLanguage,
        setCompletedStepIds,
        setStatus,
        resetTour,
    } = useQuickStartTourListState(
        useCallback(
            ({ tours, setLanguage, setCompletedStepIds, setStatus, resetTour }) => ({
                ...tours[tourId],
                setLanguage,
                setCompletedStepIds,
                setStatus,
                resetTour,
            }),
            [tourId]
        )
    )
    const onLogEvent = useCallback(
        (eventName: string, eventProperties?: any, publicArgument?: any) => {
            telemetryService.log(tourId + eventName, { language, ...eventProperties }, { language, ...publicArgument })
            console.debug(
                'DEBUG TOUR EVENT:',
                eventName,
                { language, ...eventProperties },
                { language, ...publicArgument }
            )
        },
        [language, telemetryService, tourId]
    )

    useEffect(() => {
        onLogEvent('Shown')
    }, [onLogEvent, tourId])

    const onClose = useCallback(() => {
        onLogEvent('Closed')
        setStatus(tourId, 'closed')
    }, [onLogEvent, setStatus, tourId])

    const onStepComplete = useCallback(
        (step: TourTaskStepType) => {
            const newCompletedStepIds = uniq([...completedStepIds, step.id])
            // if (completedStepIds.length === tasks.flatMap(task => task.steps).length && status !== 'closed') {
            //     setStatus('completed')
            // }
            setCompletedStepIds(tourId, newCompletedStepIds)
        },
        [completedStepIds, setCompletedStepIds, tourId]
    )

    const onStepClick = useCallback(
        (step: TourTaskStepType, language?: TourLanguage) => {
            onLogEvent(step.id + 'Clicked', { language }, { language })
            if (step.completeAfterEvents || (isLanguageRequired(step) && !language)) {
                return
            }
            onStepComplete(step)
        },
        [onLogEvent, onStepComplete]
    )

    const onLanguageSelect = useCallback(
        (language: TourLanguage) => {
            setLanguage(tourId, language)
            onLogEvent('LanguageClicked', { language }, { language })
        },
        [onLogEvent, setLanguage, tourId]
    )

    const onRestart = useCallback(
        (step: TourTaskStepType) => {
            onLogEvent(step.id + 'Clicked')
            resetTour(tourId)
        },
        [onLogEvent, resetTour, tourId]
    )

    const extendedTasks: TourTaskType[] = useMemo(
        () =>
            tasks.map(task => {
                const extendedSteps = task.steps.map(step => ({
                    ...step,
                    isCompleted: completedStepIds.includes(step.id),
                }))

                return {
                    ...task,
                    steps: extendedSteps,
                    completed: Math.round(
                        (100 * extendedSteps.filter(step => step.isCompleted).length) / extendedSteps.length
                    ),
                }
            }),
        [tasks, completedStepIds]
    )

    useEffect(() => {
        if (
            !['completed', 'closed'].includes(status as string) &&
            extendedTasks.filter(step => step.completed === 100).length === extendedTasks.length
        ) {
            onLogEvent('Completed')
            setStatus(tourId, 'completed')
        }
    }, [status, extendedTasks, onLogEvent, setStatus, tourId])

    if (status === 'closed') {
        return null
    }

    return (
        <TourContext.Provider value={{ onStepClick, language, onLanguageSelect, onRestart }}>
            <TourContent
                {...props}
                onClose={onClose}
                tasks={
                    [...extendedTasks, status === 'completed' && extraTask].filter(Boolean) as (
                        | TourTaskType
                        | TourTaskType
                    )[]
                }
            />
            <TourAgent tasks={extendedTasks} telemetryService={telemetryService} onStepComplete={onStepComplete} />
        </TourContext.Provider>
    )
}
