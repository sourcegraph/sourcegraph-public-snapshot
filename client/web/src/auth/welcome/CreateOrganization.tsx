import classNames from 'classnames'
import React, { useState, useRef, SyntheticEvent, useCallback } from 'react'

import { Button, ProductStatusBadge } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { useSteps } from '../Steps'

import styles from './CreateOrganization.module.scss'
import { useHubSpotForm } from './useHubSpotForm'

interface CreateOrganization {}

const PORTAL_ID = '2762526'
const FORM_ID = 'e0e43746-83e9-4133-97bd-9954a60c7af8'

export const CreateOrganization: React.FunctionComponent<CreateOrganization> = () => {
    const contentReference = useRef<HTMLDivElement | null>(null)
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const [isTransitioning, setIsTransitioning] = useState<boolean>(false)

    const { setComplete, currentIndex } = useSteps()

    const logFormSubmission = useCallback(() => {
        eventLogger.log('PostSignUpOrgTabBetaFormSubmit')
        setComplete(currentIndex, true)
    }, [currentIndex, setComplete])

    const form = useHubSpotForm({
        portalId: PORTAL_ID,
        formId: FORM_ID,
        onFormSubmitted: logFormSubmission,
    })

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
        <div className="mt-2 w-100">
            <h3>
                Create an organization (optional) <ProductStatusBadge status="beta" className="text-uppercase mr-1" />
            </h3>
            <p className="text-muted">
                Teams on Sourcegraph Cloud will be the quickest way to level up your team with powerful code search.
            </p>
            <div
                className={classNames({ [styles.content]: true, [styles.contentTransitioning]: isTransitioning })}
                onTransitionEnd={onTransitionEnd}
                ref={contentReference}
            >
                {!isExpanded && (
                    <div className={styles.contentInner}>
                        <p>Would you like to be added to the teams beta?</p>
                        <Button onClick={onClick} variant="success">
                            Apply
                        </Button>
                    </div>
                )}

                <div
                    className={classNames({
                        [styles.formWrapper]: true,
                        [styles.formWrapperExpanded]: isExpanded,
                    })}
                >
                    <p>Complete the form below and we’ll reach out to discuss the early beta.</p>
                    <div className={styles.form}>{form}</div>
                </div>
            </div>
        </div>
    )
}
