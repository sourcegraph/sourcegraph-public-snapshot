import classNames from 'classnames'
import React, { useState, useRef, SyntheticEvent, useCallback, useMemo } from 'react'

import { Button, ProductStatusBadge } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { FinishWelcomeFlow } from '../PostSignUpPage'
import { useSteps } from '../Steps'

import { Footer } from './Footer'
import styles from './TeamsBeta.module.scss'
import { useHubSpotForm } from './useHubSpotForm'

const PORTAL_ID = '2762526'
const FORM_ID = 'b65cc7a2-75ad-4114-be4c-cd9637e7c068'

interface TeamsBeta {
    onFinish: FinishWelcomeFlow
    onError: (error: Error) => void
}

export const TeamsBeta: React.FunctionComponent<TeamsBeta> = ({ onFinish, onError }) => {
    const contentReference = useRef<HTMLDivElement | null>(null)
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const [isTransitioning, setIsTransitioning] = useState<boolean>(false)

    const { setComplete, currentIndex } = useSteps()

    const logFormSubmission = useCallback(() => {
        eventLogger.log('PostSignUpOrgTabBetaFormSubmit')
        setComplete(currentIndex, true)
    }, [currentIndex, setComplete])

    const config = useMemo(
        () => ({
            hubSpotConfig: {
                portalId: PORTAL_ID,
                formId: FORM_ID,
            },
            onFormSubmitted: logFormSubmission,
            onError,
            initialFormValues: {},
        }),
        [logFormSubmission, onError]
    )
    const form = useHubSpotForm(config)

    function onClick(): void {
        eventLogger.log('PostSignUpOrgTabApplyToBeta')
        setIsTransitioning(true)
    }

    function onTransitionEnd(event: SyntheticEvent): void {
        if (event.target === contentReference.current) {
            setIsExpanded(true)
        }
    }

    return (
        <div className={classNames('mt-3', styles.container)}>
            <div className="mb-3">
                <h3>
                    Apply to the team beta (optional){' '}
                    <ProductStatusBadge status="beta" className="text-uppercase mr-1" />
                </h3>
                <p className="text-muted mb-0">
                    Teams on Sourcegraph Cloud will be the quickest way to level up your team with powerful code search.
                </p>
            </div>
            <div
                className={classNames({ [styles.content]: true, [styles.contentTransitioning]: isTransitioning })}
                onTransitionEnd={onTransitionEnd}
                ref={contentReference}
            >
                {!isExpanded && (
                    <div className={styles.contentInner}>
                        <p>Click the button below to apply to the beta.</p>
                        <Button onClick={onClick} variant="success">
                            Apply to beta
                        </Button>
                    </div>
                )}

                <div
                    className={classNames({
                        [styles.form]: true,
                        [styles.formExpanded]: isExpanded,
                    })}
                >
                    {form}
                </div>
            </div>
            <Footer onFinish={onFinish} isSkippable={true} />
        </div>
    )
}
