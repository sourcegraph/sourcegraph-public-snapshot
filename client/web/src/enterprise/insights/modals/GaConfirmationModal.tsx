import React, { useContext, useMemo } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Modal, Link } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'

import { FourLineChart, LangStatsInsightChart, ThreeLineChart } from './components/MediaCharts'

import styles from './GaConfirmationModal.module.scss'

export const GaConfirmationModal: React.FunctionComponent = () => {
    const [isGaAccepted, setGaAccepted] = useTemporarySetting('insights.freeGaAccepted', false)
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)
    const features = useMemo(() => getUiFeatures(), [getUiFeatures])

    const showConfirmationModal = !features?.licensed && isGaAccepted === false

    if (!showConfirmationModal) {
        return null
    }

    const handleAccept = (): void => {
        setGaAccepted(true)
    }

    return (
        <Modal position="center" aria-label="Code Insights Ga information" containerClassName={styles.overlay}>
            <GaConfirmationModalContent onAccept={handleAccept} />
        </Modal>
    )
}

interface GaConfirmationModalContentProps {
    onAccept: () => void
}

/**
 * Renders Code Insights Ga modal content component.
 * Exported especially for storybook story component cause chromatic has a problem of rendering modals
 * on CI.
 */
export const GaConfirmationModalContent: React.FunctionComponent<GaConfirmationModalContentProps> = props => {
    const { onAccept } = props

    return (
        <>
            <h1 className={styles.title}>Thank you for trying Code Insights!</h1>

            <div className={styles.mediaHeroWrapper}>
                <div className={styles.mediaHeroContent}>
                    <ThreeLineChart className={styles.chart} />
                    <FourLineChart className={styles.chart} />
                    <LangStatsInsightChart className={styles.chart} />
                </div>
                <div className={styles.mediaHeroOverlay}>Your trial has expired</div>
            </div>

            <div className={styles.textContent}>
                <p>
                    <b>Your instance is now using the limited access version of Code Insights.</b> This version is
                    limited to 2 code insights.
                </p>

                <p>
                    If your team has created more than 2 insights during the trial, some insights will be hidden. All of
                    the insights you've created are preserved, <b>but</b> will be locked. All insights will be visible
                    again after upgrading your license.
                </p>

                <p>
                    Contact your admin or reach out to us to upgrade your licence for unlimited insights and dashboards.
                </p>

                <p>
                    Questions? Learn more about the{' '}
                    <Link to="/help/code_insights/references/license">Code Insights limited access</Link> or contact us
                    directly.
                </p>
            </div>

            <footer className={styles.actions}>
                <Button variant="primary" onClick={onAccept}>
                    Understood, let’s go!
                </Button>
            </footer>
        </>
    )
}
