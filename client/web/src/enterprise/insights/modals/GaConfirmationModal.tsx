import React, { useContext, useMemo, useRef } from 'react'
import { useHistory } from 'react-router'
import { Observable } from 'rxjs'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, useAutoFocus, Modal, Link, useObservable } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'

import { FourLineChart, LangStatsInsightChart, ThreeLineChart } from './components/MediaCharts'
import styles from './GaConfirmationModal.module.scss'

export interface GaConfirmationModalProps {
    fetchSiteUpdateCheck: () => Observable<{ productVersion: string }>
}

export const GaConfirmationModal: React.FunctionComponent<GaConfirmationModalProps> = ({ fetchSiteUpdateCheck }) => {
    const history = useHistory()
    const [isGaAccepted, setGaAccepted] = useTemporarySetting('insights.freeGaAccepted', false)
    const { isCodeInsightsLicensed } = useContext(CodeInsightsBackendContext)
    const site = useObservable(useMemo(() => fetchSiteUpdateCheck(), [fetchSiteUpdateCheck]))
    const isLicensed = useObservable(useMemo(() => isCodeInsightsLicensed(), [isCodeInsightsLicensed]))

    const isSourcegraph3_37_x = site?.productVersion.startsWith('3.37')

    const showConfirmationModal = isSourcegraph3_37_x && !isLicensed && !isGaAccepted

    if (!showConfirmationModal) {
        return null
    }

    const handleAccept = (): void => {
        setGaAccepted(true)
    }

    const handleDismiss = (): void => {
        history.push('/')
    }

    return (
        <Modal position="center" aria-label="Code Insights Ga information" containerClassName={styles.overlay}>
            <GaConfirmationModalContent onAccept={handleAccept} onDismiss={handleDismiss} />
        </Modal>
    )
}

interface GaConfirmationModalContentProps {
    onAccept: () => void
    onDismiss: () => void
}

/**
 * Renders Code Insights Ga modal content component.
 * Exported especially for storybook story component cause chromatic has a problem of rendering modals
 * on CI.
 */
export const GaConfirmationModalContent: React.FunctionComponent<GaConfirmationModalContentProps> = props => {
    const { onAccept, onDismiss } = props
    const dismissButtonReference = useRef<HTMLButtonElement>(null)

    useAutoFocus({ autoFocus: true, reference: dismissButtonReference })

    return (
        <>
            <h1 className={styles.title}>Code Insights is now Generally Available</h1>

            <div className={styles.mediaHeroContent}>
                <ThreeLineChart className={styles.chart} />
                <FourLineChart className={styles.chart} />
                <LangStatsInsightChart className={styles.chart} />
            </div>

            <div className={styles.textContent}>
                <p>Code Insights are officially out of beta!</p>

                <p>
                    You will keep full access to Code Insights while on this Sourcegraph version as a Free Trial. After
                    this period, you will either need to purchase Code Insights to continue full functionality or you
                    will only be able to use a limited number of Code Insights.
                </p>

                <p>
                    Contact your admin or reach out to us to upgrade your licence for unlimited insights and dashboards.
                </p>

                <p>
                    Questions? Learn more about the <Link to="#">Code Insights limited access</Link> or{' '}
                    <Link to="mailto:support@sourcegraph.com">contact us directly</Link>.
                </p>
            </div>

            <footer className={styles.actions}>
                <Button ref={dismissButtonReference} variant="secondary" outline={true} onClick={onDismiss}>
                    Maybe later
                </Button>

                <Button variant="primary" onClick={onAccept}>
                    Understood, let’s go!
                </Button>
            </footer>
        </>
    )
}
