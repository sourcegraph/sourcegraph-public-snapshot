import { FunctionComponent, useEffect, useState } from 'react'

import { Button, H2, H3, Text, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'

import { Illustration } from './Illustration'

import styles from './TrySourcegraphCloudSection.module.scss'

export const TrySourcegraphCloudSection: FunctionComponent = () => {
    const bylines: string[] = [
        'Understand and search.',
        'Navigate your codebase.',
        'Automate large-scale code changes.',
        'Get insights.',
    ]
    const [bylineIndex, setBylineIndex] = useState(0)
    const [byline, setByline] = useState(bylines[0])

    useEffect(() => {
        /**
         * Base metric for when the section loads which we can use to count
         * against impressions/views vs. clicks.
         */
        eventLogger.log('DisplayOfCloudCTA')

        const cycle = setInterval(() => {
            const newIndex = bylineIndex === bylines.length - 1 ? 0 : bylineIndex + 1

            setByline(bylines[newIndex])
            setBylineIndex(newIndex)
        }, 3500)

        return () => clearInterval(cycle)
    })

    return (
        <div>
            <Link
                to="https://signup.sourcegraph.com/"
                onClick={() => eventLogger.log('ClickedOnCloudCTA')}
                className={styles.wrapper}
            >
                <Illustration className={styles.illustration} />

                <div className={styles.content}>
                    <div>
                        <H2 className={styles.tryCloud}>Try Sourcegraph Cloud.</H2>
                        <H3 className={styles.byline}>{byline}</H3>
                        <Text className={styles.signUp}>Sign up for a 30-day trial for your team.</Text>
                    </div>

                    <div className={styles.buttonContainer}>
                        <Button
                            variant="secondary"
                            className={styles.trialButton}
                        >
                            Get free trial now
                        </Button>
                    </div>
                </div>
            </Link>
            
            <Text size="small" className={styles.selfHostedCopy}>
                Want to deploy yourself?
                
                {/* eslint-disable-next-line react/forbid-elements */}
                <Link
                    to="/help"
                    onClick={() => eventLogger.log('ClickedOnSelfHostedCTA')}
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Try our self-hosted solution.
                </Link>
            </Text>
        </div>
    )
}
