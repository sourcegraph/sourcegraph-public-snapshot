import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { TelemetryProps, TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Button, ButtonLink } from '@sourcegraph/wildcard'

import { OnboardingTourLanguage, useOnboardingTourState } from '../stores/onboardingTourState'

import { OnboardingTourStepItem, ONBOARDING_STEP_ITEMS } from './data'
import styles from './OnboardingTour.module.scss'
import { OnboardingTourSteps } from './OnboardingTourSteps'
import { useLogTourEvent } from './useOnboardingTour'
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

interface LanguagePickerProps extends TelemetryProps {
    steps: OnboardingTourStepItem[]
}

const LanguagePicker: React.FunctionComponent<LanguagePickerProps> = ({ steps, telemetryService }) => {
    const logTourEvent = useLogTourEvent(telemetryService)
    const { continueID, addCompletedID, setLanguage } = useOnboardingTourState(
        useCallback(({ continueID, addCompletedID, setLanguage }) => ({ continueID, addCompletedID, setLanguage }), [])
    )
    const history = useHistory()
    const createOnClickHandler = useCallback(
        (language: OnboardingTourLanguage) => () => {
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
                {Object.values(OnboardingTourLanguage).map(language => (
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

const TourComplete: React.FunctionComponent<TelemetryProps> = ({ telemetryService }) => {
    const logTourEvent = useLogTourEvent(telemetryService)
    const restart = useOnboardingTourState(useCallback(state => state.restart, []))

    const onGetStarted = useCallback(() => {
        logTourEvent('TourGetStartedClicked')
    }, [logTourEvent])

    return (
        <>
            <p className={styles.text}>
                Install Sourcegraph locally or create an account to get powerful code search and other powerful features
                on your private code.
            </p>
            <div className="d-flex flex-column">
                <ButtonLink
                    className="align-self-start mb-2"
                    to={buildGetStartedURL('onboarding-tour')}
                    onClick={onGetStarted}
                    variant="primary"
                >
                    Get started
                </ButtonLink>
                <Button variant="link" size="sm" className="align-self-start text-left pl-0" onClick={restart}>
                    Restart
                </Button>
            </div>
        </>
    )
}

function useSteps(telemetryService: TelemetryService): OnboardingTourStepItem[] {
    const logTourEvent = useLogTourEvent(telemetryService)
    const { status, complete, completedIDs } = useOnboardingTourState(
        useCallback(({ status, complete, completedIDs }) => ({ status, complete, completedIDs }), [])
    )

    const steps = useMemo(
        () =>
            ONBOARDING_STEP_ITEMS.map(step => ({
                ...step,
                isCompleted: !!completedIDs?.includes(step.id),
            })),
        [completedIDs]
    )

    useEffect(() => {
        if (status !== 'completed' && steps.filter(step => step.isCompleted).length === steps.length) {
            logTourEvent('TourComplete')
            complete()
        }
    }, [complete, status, steps, logTourEvent])

    return steps
}

export interface OnboardingTourManagerProps extends TelemetryProps {
    isFixedHeight?: boolean
    className?: string
}

export const OnboardingTourManager: React.FunctionComponent<OnboardingTourManagerProps> = ({
    className,
    isFixedHeight,
    telemetryService,
}) => {
    const steps = useSteps(telemetryService)
    const { close, status } = useOnboardingTourState(useCallback(({ close, status }) => ({ close, status }), []))

    if (status === 'closed') {
        return null
    }

    return (
        <Card
            title={status === 'completed' ? 'Tour complete!' : 'Getting Started'}
            onClose={close}
            showDivider={status !== 'completed'}
            className={className}
        >
            {status === 'steps' ? (
                // Main tour steps
                <OnboardingTourSteps
                    steps={steps}
                    telemetryService={telemetryService}
                    className={classNames({ [styles.isFixedHeight]: isFixedHeight })}
                />
            ) : status === 'languages' ? (
                // Pick language for the tour
                <LanguagePicker steps={steps} telemetryService={telemetryService} />
            ) : (
                // Sign-up or restart
                <TourComplete telemetryService={telemetryService} />
            )}
        </Card>
    )
}
