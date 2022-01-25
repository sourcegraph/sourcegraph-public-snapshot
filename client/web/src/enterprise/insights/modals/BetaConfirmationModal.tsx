import React, { useRef } from 'react'
import { useHistory } from 'react-router'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, useAutoFocus, Modal, Link } from '@sourcegraph/wildcard'

import styles from './BetaConfirmationModal.module.scss'
import { FourLineChart, LangStatsInsightChart, ThreeLineChart } from './components/MediaCharts'

export const BetaConfirmationModal: React.FunctionComponent = () => {
    const history = useHistory()
    const [isFreeBetaAccepted, setFreeBetaAccepted] = useTemporarySetting('insights.freeBetaAccepted', false)

    const handleAccept = (): void => {
        setFreeBetaAccepted(true)
    }

    const handleDismiss = (): void => {
        history.push('/')
    }

    // We should not render confirmation modal if we haven't got the temporary settings yet
    // or cause users have already accepted the free beta info.
    if (isFreeBetaAccepted === undefined || isFreeBetaAccepted) {
        return null
    }

    return (
        <Modal position="center" aria-label="Code Insights Beta information" containerClassName={styles.overlay}>
            <BetaConfirmationModalContent onAccept={handleAccept} onDismiss={handleDismiss} />
        </Modal>
    )
}

interface BetaConfirmationModalContentProps {
    onAccept: () => void
    onDismiss: () => void
}

/**
 * Renders Code Insights Beta modal content component.
 * Exported especially for storybook story component cause chromatic has a problem of rendering modals
 * on CI.
 */
export const BetaConfirmationModalContent: React.FunctionComponent<BetaConfirmationModalContentProps> = props => {
    const { onAccept, onDismiss } = props
    const dismissButtonReference = useRef<HTMLButtonElement>(null)

    useAutoFocus({ autoFocus: true, reference: dismissButtonReference })

    return (
        <>
            <h1 className={styles.title}>Welcome to the Code Insights Beta!</h1>

            <div className={styles.mediaHeroContent}>
                <ThreeLineChart className={styles.chart} />
                <FourLineChart className={styles.chart} />
                <LangStatsInsightChart className={styles.chart} />
            </div>

            <div className={styles.textContent}>
                <p>
                    <b>🥁 We’re introducing Code Insights</b>: a new analytics tool that lets you track and understand
                    what’s in your code and how it changes <b>over time</b>.
                </p>

                <p>
                    Track anything that can be expressed with a Sourcegraph search query: migrations, package use,
                    version adoption, code smells, codebase size, and more, across 1,000s of repositories.
                </p>

                <p>
                    We're still polishing Code Insights and you might find bugs while we’re in beta. Please{' '}
                    <Link
                        to="https://docs.sourcegraph.com/code_insights#code-insights-beta"
                        target="_blank"
                        rel="noopener"
                    >
                        share any bugs 🐛 or feedback
                    </Link>{' '}
                    to help us make Code Insights better.
                </p>

                <p>
                    Code Insights is <b>free while in beta through 2021</b>. When Code Insights is officially released,
                    we may disable your use of the product or charge for continued use.
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
