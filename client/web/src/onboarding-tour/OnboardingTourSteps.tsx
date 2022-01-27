import { Accordion, AccordionButton, AccordionItem, AccordionPanel } from '@reach/accordion'
import classNames from 'classnames'
import { groupBy } from 'lodash'
import ArrowDropDownIcon from 'mdi-react/ArrowDropDownIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { CircularProgressbar } from 'react-circular-progressbar'
import { Link, useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { useOnboardingTourState } from '../stores/onboardingTourState'

import { OnboardingTourStepItem } from './data'
import styles from './OnboardingTour.module.scss'
import { OnboardingTourInfoAgent, OnboardingTourCompletionAgent } from './OnboardingTourAgents'
import { useLogTourEvent } from './useOnboardingTour'
import { buildURIMarkers, isExternalURL, parseURIMarkers } from './utils'

interface LinkOrAnchorProps {
    href: string
    className?: string
    onClick?: (event: React.MouseEvent<HTMLAnchorElement>) => void
}

const LinkOrAnchor: React.FunctionComponent<LinkOrAnchorProps> = ({ href, children, ...props }) => {
    if (isExternalURL(href)) {
        return (
            <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
                {children}
            </a>
        )
    }

    return (
        <Link to={href} {...props}>
            {children}
        </Link>
    )
}

interface OnboardingTourStepProps extends OnboardingTourStepItem, TelemetryProps {}

const OnboardingTourStep: React.FunctionComponent<OnboardingTourStepProps> = ({
    completeAfterEvents,
    isCompleted,
    url,
    title,
    id,
    telemetryService,
}) => {
    const logTourEvent = useLogTourEvent(telemetryService)
    const { language, addCompletedID, setLanguageStatus } = useOnboardingTourState(
        useCallback(
            ({ language, addCompletedID, setLanguageStatus }) => ({ language, addCompletedID, setLanguageStatus }),
            []
        )
    )
    const onClick = useCallback(
        (event: React.MouseEvent<HTMLAnchorElement>) => {
            if (!language && !isExternalURL(event.currentTarget.href)) {
                setLanguageStatus(id)
                event.preventDefault()
            }
            logTourEvent(id + 'Clicked')
            if (!completeAfterEvents) {
                addCompletedID(id)
            }
        },
        [language, logTourEvent, id, completeAfterEvents, setLanguageStatus, addCompletedID]
    )

    const href = typeof url === 'string' ? url : language ? url[language] : '#'

    return (
        <div className={styles.step}>
            <LinkOrAnchor
                onClick={onClick}
                href={buildURIMarkers(href, id)}
                className={classNames(styles.label, styles.link)}
            >
                {title}
            </LinkOrAnchor>
            <CheckCircleIcon
                className={classNames('icon-inline', isCompleted ? 'text-success' : styles.iconMuted)}
                size="1rem"
            />
        </div>
    )
}
interface OnboardingTourStepsProps extends TelemetryProps {
    className?: string
    steps: OnboardingTourStepItem[]
}
export const OnboardingTourSteps: React.FunctionComponent<OnboardingTourStepsProps> = ({
    steps,
    className,
    telemetryService,
}) => {
    const location = useLocation()
    const [expandedIndex, setExpandedIndex] = useState<number[]>([])
    const groups = useMemo(
        () =>
            Object.entries(groupBy(steps, 'group')).map(([title, steps]) => ({
                title,
                steps,
                completed: Math.round((100 * steps.filter(step => step.isCompleted).length) / steps.length),
            })),
        [steps]
    )
    const completedCount = useMemo(() => steps.filter(step => step.isCompleted).length, [steps])

    const toggleExpandedIndexes = useCallback((currentIndex: number) => {
        setExpandedIndex(indexes => {
            if (indexes.includes(currentIndex)) {
                return indexes.filter(index => index !== currentIndex)
            }

            return [...indexes, currentIndex]
        })
    }, [])

    useEffect(() => {
        const { stepId } = parseURIMarkers(location.search)
        const currentIndex = groups.findIndex(group => group.steps.find(step => step.id === stepId))
        if (currentIndex >= 0) {
            setExpandedIndex(indexes => [...indexes, currentIndex])
        }
    }, [location, groups])

    return (
        <>
            <Accordion className={className} index={expandedIndex} onChange={toggleExpandedIndexes}>
                {groups.map(({ title, steps, completed }) => (
                    <AccordionItem key={title}>
                        <AccordionButton as="div">
                            <span className={styles.arrowIconContainer}>
                                <ArrowDropDownIcon size="1rem" />
                            </span>
                            <span className={styles.label}>{title}</span>
                            {completed < 100 ? (
                                <CircularProgressbar
                                    className={styles.progressBar}
                                    strokeWidth={20}
                                    value={completed}
                                />
                            ) : (
                                <CheckCircleIcon className={classNames('icon-inline', 'text-success')} size="1rem" />
                            )}
                        </AccordionButton>
                        <AccordionPanel>
                            {steps.map(step => (
                                <OnboardingTourStep key={step.id} telemetryService={telemetryService} {...step} />
                            ))}
                        </AccordionPanel>
                    </AccordionItem>
                ))}
            </Accordion>

            <footer>
                <CheckCircleIcon className="icon-inline text-success" size="1rem" />
                <span className={styles.footerText}>
                    {completedCount} of {steps.length}
                </span>
            </footer>
            <OnboardingTourInfoAgent steps={steps} />
            <OnboardingTourCompletionAgent steps={steps} telemetryService={telemetryService} />
        </>
    )
}
