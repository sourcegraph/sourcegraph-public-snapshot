import { DialogContent, DialogOverlay } from '@reach/dialog'
import React from 'react'
import { useHistory } from 'react-router'

import { Button } from '@sourcegraph/wildcard'

import { useTemporarySetting } from '../../../settings/temporary/useTemporarySetting'

import styles from './BetaConfirmationModal.module.scss'
import { FourLineChart, ThreeLineChart, TwoLineChart } from './components/MediaCharts'

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
                <h1 className={styles.title}>Welcome to Code Insights Free Beta!</h1>

                <div className={styles.mediaHeroContent}>
                    <ThreeLineChart className={styles.chart} />
                    <FourLineChart className={styles.chart} />
                    <TwoLineChart className={styles.chart} />
                </div>

                <div className={styles.textContent}>
                    <p>
                        <b>🥁 We’re introducing Code Insights</b>, the first code analytics tool that can tell you
                        things about your code base at a <b>high level</b>!
                    </p>

                    <p>
                        Code Insights are based on our universal code search, making them <b>incredibly accurate</b>.
                        Track anything that can be expressed with Sourcegraph search query: migrations, usage of
                        packages and much much more.
                    </p>

                    <p>
                        We're still polishing Insights and you may experience some issues while in beta.{' '}
                        <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                            Share any bugs 🐛 or feedback
                        </a>{' '}
                        to help us make Code Insights better.
                    </p>

                    <p>
                        Code Insights are <b>free and in beta through 2021</b>. In 2022, Code Insights may be included
                        in a separate paid plan.
                    </p>
                </div>

                <footer className={styles.actions}>
                    <Button variant="secondary" outline={true} onClick={() => history.push('/')}>
                        Maybe later
                    </Button>

                    <Button variant="primary" onClick={handleAcceptClick}>
                        Understood, let's try it!
                    </Button>
                </footer>
            </DialogContent>
        </DialogOverlay>
    )
}
