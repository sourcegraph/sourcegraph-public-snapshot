import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { Accordion, AccordionButton, AccordionItem, AccordionPanel } from '@reach/accordion'
import classNames from 'classnames'
import { groupBy } from 'lodash'
import ArrowDropDownIcon from 'mdi-react/ArrowDropDownIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import { CircularProgressbar } from 'react-circular-progressbar'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon, Link } from '@sourcegraph/wildcard'

import { useGettingStartedTourState } from '../stores/gettingStartedTourState'

import { GettingStartedTourStepItem } from './data'
import { GettingStartedTourInfoAgent, GettingStartedTourCompletionAgent } from './GettingStartedTourAgents'
import { useGettingStartedTourLogEvent } from './useGettingStartedTourLogEvent'
import { buildURIMarkers, isExternalURL, parseURIMarkers } from './utils'

import styles from './GettingStartedTour.module.scss'

interface LinkOrAnchorProps {
    href: string
    className?: string
    onClick?: (event: React.MouseEvent<HTMLAnchorElement>) => void
}

const LinkOrAnchor: React.FunctionComponent<LinkOrAnchorProps> = ({ href, children, ...props }) => (
    <Link to={href} {...props} {...(isExternalURL(href) && { target: '_blank', rel: 'noopener noreferrer' })}>
        {children}
    </Link>
)

interface GettingStartedTourStepProps extends GettingStartedTourStepItem, TelemetryProps {}

const GettingStartedTourStep: React.FunctionComponent<GettingStartedTourStepProps> = ({
    completeAfterEvents,
    isCompleted,
    url,
    title,
    id,
    telemetryService,
}) => {
    const logTourEvent = useGettingStartedTourLogEvent(telemetryService)
    const { language, addCompletedID, setLanguageStatus } = useGettingStartedTourState(
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
            <Icon className={classNames(isCompleted ? 'text-success' : styles.iconMuted)} as={CheckCircleIcon} />
        </div>
    )
}
interface GettingStartedTourStepsProps extends TelemetryProps {
    className?: string
    steps: GettingStartedTourStepItem[]
}
export const GettingStartedTourSteps: React.FunctionComponent<GettingStartedTourStepsProps> = ({
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
                                <Icon className="text-success" as={CheckCircleIcon} />
                            )}
                        </AccordionButton>
                        <AccordionPanel>
                            {steps.map(step => (
                                <GettingStartedTourStep key={step.id} telemetryService={telemetryService} {...step} />
                            ))}
                        </AccordionPanel>
                    </AccordionItem>
                ))}
            </Accordion>

            <footer>
                <Icon className="text-success" as={CheckCircleIcon} />
                <span className={styles.footerText}>
                    {completedCount} of {steps.length}
                </span>
            </footer>
            <GettingStartedTourInfoAgent steps={steps} />
            <GettingStartedTourCompletionAgent steps={steps} telemetryService={telemetryService} />
        </>
    )
}
