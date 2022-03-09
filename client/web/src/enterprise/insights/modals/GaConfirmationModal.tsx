import React, { useContext, useMemo } from 'react'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Modal, Link, useObservable } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'

import { FourLineChart, LangStatsInsightChart, ThreeLineChart } from './components/MediaCharts'
import styles from './GaConfirmationModal.module.scss'

export const GaConfirmationModal: React.FunctionComponent = () => {
    const [isGaAccepted, setGaAccepted] = useTemporarySetting('insights.freeGaAccepted', false)
    const { getUiFeatures } = useContext(CodeInsightsBackendContext)
    const features = useObservable(useMemo(() => getUiFeatures(), [getUiFeatures]))

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
            <h1 className={styles.title}>Code Insights is Generally Available</h1>

            <div className={styles.mediaHeroContent}>
                <ThreeLineChart className={styles.chart} />
                <FourLineChart className={styles.chart} />
                <LangStatsInsightChart className={styles.chart} />
            </div>

            <div className={styles.textContent}>
                <p>
                    Code Insights is a new analytics product that transforms your code into a queryable database so you
                    can create customizable, visual dashboards to understand your codebase at a high level.
                </p>

                <p>
                    Code Insights is now Generally Available.{' '}
                    <b>You can create unlimited insights and dashboards while on this version of Sourcegraph.</b> In the
                    next version upgrade, you will either need to purchase Code Insights to continue using its full
                    functionality, or you can use{' '}
                    <Link to="/help/code_insights/references/license">a limited number of insights for free</Link>.
                </p>

                <p>
                    Reach out to your Sourcegraph admin or account team to purchase Code Insights. Questions? Please{' '}
                    <Link to="mailto:support@sourcegraph.com">contact us</Link>.
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
