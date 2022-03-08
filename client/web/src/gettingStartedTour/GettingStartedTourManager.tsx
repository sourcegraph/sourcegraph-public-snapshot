import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Button, ButtonLink } from '@sourcegraph/wildcard'

import {
    GettingStartedTourLanguage,
    GettingStartedTourState,
    useGettingStartedTourState,
} from '../stores/gettingStartedTourState'

import { GettingStartedTourStepItem, GETTING_STARTED_TOUR_STEP_ITEMS } from './data'
import styles from './GettingStartedTour.module.scss'
import { GettingStartedTourSteps } from './GettingStartedTourSteps'
import { useGettingStartedTourLogEvent } from './useGettingStartedTourLogEvent'
import { buildURIMarkers, isExternalURL } from './utils'

interface CardProps {
    onClose: () => void
    title: string
    className?: string
    showDivider?: boolean
}

const Card: React.FunctionComponent<CardProps> = ({ title, children, onClose, className, showDivider }) => (
    <article className={classNames(styles.card, className)}>
        <div className={styles.cardInner}>
            <header className={styles.cardHeader}>
                <h3 className={classNames(styles.cardTitle)}>{title}</h3>
                <CloseIcon onClick={onClose} size="1rem" />
            </header>
            {showDivider && <hr className={styles.divider} />}
            <div className="flex-grow-1">{children}</div>
        </div>
    </article>
)

interface GettingStartedTourLanguagePickerProps extends TelemetryProps {
    steps: GettingStartedTourStepItem[]
}

const GettingStartedTourLanguagePicker: React.FunctionComponent<GettingStartedTourLanguagePickerProps> = ({
    steps,
    telemetryService,
}) => {
    const logTourEvent = useGettingStartedTourLogEvent(telemetryService)
    const { continueID, addCompletedID, setLanguage } = useGettingStartedTourState(
        useCallback(({ continueID, addCompletedID, setLanguage }) => ({ continueID, addCompletedID, setLanguage }), [])
    )
    const history = useHistory()
    const createOnClickHandler = useCallback(
        (language: GettingStartedTourLanguage) => () => {
            setLanguage(language)
            const step = steps.find(step => step.id === continueID)
            if (!step) {
                return
            }
            const url = typeof step.url === 'string' ? step.url : step.url[language]
            if (isExternalURL(url)) {
                window.open(url, '_blank')
            } else {
                history.push(buildURIMarkers(url, step.id))
            }
            addCompletedID(step.id)
            logTourEvent(step.id + 'Clicked')
        },
        [setLanguage, steps, addCompletedID, logTourEvent, continueID, history]
    )

    return (
        <>
            <p className={classNames(styles.languageTitle, 'mt-2')}>
                This guide is available in the following languages:
            </p>
            <div className="d-flex flex-wrap mt-3 mb-1">
                {Object.values(GettingStartedTourLanguage).map(language => (
                    <Button
                        key={language}
                        onClick={createOnClickHandler(language)}
                        size="sm"
                        className={classNames('mr-1 my-1', styles.language)}
                    >
                        {language}
                    </Button>
                ))}
            </div>
        </>
    )
}

const GettingStartedTourComplete: React.FunctionComponent<TelemetryProps> = ({ telemetryService }) => {
    const logTourEvent = useGettingStartedTourLogEvent(telemetryService)
    const restart = useGettingStartedTourState(useCallback(state => state.restart, []))

    const onGetStarted = useCallback(() => {
        logTourEvent('TourGetStartedClicked')
    }, [logTourEvent])

    const onRestart = useCallback(() => {
        logTourEvent('TourRestartClicked')
        restart()
    }, [logTourEvent, restart])

    return (
        <>
            <p className={styles.text}>
                Install Sourcegraph locally or create an account to get powerful code search and other powerful features
                on your private code.
            </p>
            <div className="d-flex flex-column">
                <ButtonLink
                    className="align-self-start mb-2"
                    to={buildGetStartedURL('getting-started-tour')}
                    onClick={onGetStarted}
                    variant="primary"
                >
                    Get started
                </ButtonLink>
                <Button variant="link" size="sm" className="align-self-start text-left pl-0" onClick={onRestart}>
                    Restart
                </Button>
            </div>
        </>
    )
}

function useGettingStartedTourManager(
    telemetryService: TelemetryService
): { steps: GettingStartedTourStepItem[]; onClose: () => void; status: GettingStartedTourState['status'] } {
    const logTourEvent = useGettingStartedTourLogEvent(telemetryService)
    const { status, complete, completedIDs, close } = useGettingStartedTourState(
        useCallback(({ status, complete, completedIDs, close }) => ({ status, complete, completedIDs, close }), [])
    )

    const steps = useMemo(
        () =>
            GETTING_STARTED_TOUR_STEP_ITEMS.map(step => ({
                ...step,
                isCompleted: !!completedIDs?.includes(step.id),
            })),
        [completedIDs]
    )

    // Handle on complete
    useEffect(() => {
        if (
            !['completed', 'closed'].includes(status as string) &&
            steps.filter(step => step.isCompleted).length === steps.length
        ) {
            logTourEvent('TourComplete')
            complete()
        }
    }, [complete, status, steps, logTourEvent])

    // Handle on initial view
    useEffect(() => {
        if (status !== 'closed') {
            logTourEvent('TourShown')
        }
        // NOTE: intentionally excluding status to fire only once
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const onClose = useCallback(() => {
        logTourEvent('TourClosed')
        close()
    }, [close, logTourEvent])

    return { steps, onClose, status }
}

export interface GettingStartedTourManagerProps extends TelemetryProps {
    isFixedHeight?: boolean
    className?: string
}

export const GettingStartedTourManager: React.FunctionComponent<GettingStartedTourManagerProps> = ({
    className,
    isFixedHeight,
    telemetryService,
}) => {
    const { steps, onClose, status } = useGettingStartedTourManager(telemetryService)

    if (status === 'closed') {
        return null
    }

    return (
        <Card
            title={status === 'completed' ? 'Tour complete!' : 'Getting Started'}
            onClose={onClose}
            showDivider={status !== 'completed'}
            className={className}
        >
            {status === 'steps' ? (
                // Main tour steps
                <GettingStartedTourSteps
                    steps={steps}
                    telemetryService={telemetryService}
                    className={classNames({ [styles.isFixedHeight]: isFixedHeight })}
                />
            ) : status === 'languages' ? (
                // Pick language for the tour
                <GettingStartedTourLanguagePicker steps={steps} telemetryService={telemetryService} />
            ) : (
                // Sign-up or restart
                <GettingStartedTourComplete telemetryService={telemetryService} />
            )}
        </Card>
    )
}
