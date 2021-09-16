import { DialogContent, DialogOverlay } from '@reach/dialog'
import React from 'react'
import { useHistory } from 'react-router'

import { Button } from '@sourcegraph/wildcard'

import { useTemporarySetting } from '../../../settings/temporary/useTemporarySetting'

import styles from './BetaConfirmationModal.module.scss'
import { FourLineChart, PieChart, ThreeLineChart } from './components/MediaCharts'

export const BetaConfirmationModal: React.FunctionComponent = () => {
    const history = useHistory()
    const [acceptFreeBetaValue, setAcceptFreeBeta] = useTemporarySetting('insights.acceptFreeBeta', false)

    const handleAcceptClick = (): void => {
        setAcceptFreeBeta(true)
    }

    // We should not render confirmation modal if we haven't got the temporary settings yet
    // or cause users have already accepted the free beta info.
    if (acceptFreeBetaValue === undefined || acceptFreeBetaValue) {
        return null
    }

    return (
        <DialogOverlay className={styles.overlay}>
            <DialogContent className={styles.content}>
                <h1 className={styles.title}>Welcome to the Code Insights Beta!</h1>

                <div className={styles.mediaHeroContent}>
                    <ThreeLineChart className={styles.chart} />
                    <FourLineChart className={styles.chart} />
                    <PieChart className={styles.chart} />
                </div>

                <div className={styles.textContent}>
                    <p>
                        <b>🥁 We’re introducing Code Insights</b>: a new analytics tool that lets you track and
                        understand what’s in your code and how it changes <b>over time</b>!
                    </p>

                    <p>
                        Track anything that can be expressed with a Sourcegraph search query: migrations, package use,
                        version adoption, code smells, codebase size, and more, across 1,000s of repositories.
                    </p>

                    <p>
                        We're still polishing Code Insights and you might find bugs while we’re in beta. Please{' '}
                        <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                            share any bugs 🐛 or feedback
                        </a>{' '}
                        to help us make Code Insights better.
                    </p>

                    <p>
                        Code Insights is <b>free and in beta through 2021</b>. When Code Insights is officially
                        released, continued use may require a separate paid plan (at which time we’d notify you again).
                    </p>
                </div>

                <footer className={styles.actions}>
                    <Button variant="secondary" outline={true} onClick={() => history.push('/')}>
                        Maybe later
                    </Button>

                    <Button variant="primary" onClick={handleAcceptClick}>
                        Understood, let’s go!
                    </Button>
                </footer>
            </DialogContent>
        </DialogOverlay>
    )
}
