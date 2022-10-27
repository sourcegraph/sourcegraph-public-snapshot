import { FunctionComponent, useEffect } from 'react'

import { Button, H2, H3, Text, Link } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'

import styles from './TrySourcegraphCloudSection.module.scss'

export const TrySourcegraphCloudSection: FunctionComponent = () => {
    useEffect(() => {
        /**
         * Base metric for when the section loads which we can use to count
         * against impressions/views vs. clicks.
         */
        eventLogger.log('DisplayOfCloudCTA')
    }, [])

    return (
        <div>
            <Link
                to="https://signup.sourcegraph.com/"
                onClick={() => eventLogger.log('ClickedOnCloudCTA')}
                className={styles.wrapper}
            >
                <img
                    src="https://storage.googleapis.com/sourcegraph-assets/search/homepage/illustration.svg"
                    alt="abstract triangles"
                    className={styles.illustration}
                />

                <div className={styles.content}>
                    <div>
                        <H2 className={styles.tryCloud}>Try Sourcegraph Cloud.</H2>
                        <H3 className={styles.bylines}>
                            <span>Understand and search.</span>
                            <span>Navigate your codebase.</span>
                            <span>Automate large-scale code changes.</span>
                            <span>Get insights.</span>
                            <span>Understand and search.</span>
                        </H3>
                        <Text className={styles.signUp}>Sign up for a 30-day trial for your team.</Text>
                    </div>

                    <div className={styles.buttonContainer}>
                        <Button variant="secondary" className={styles.trialButton}>
                            Get free trial now
                        </Button>
                    </div>
                </div>
            </Link>

            <Text size="small" className={styles.selfHostedCopy}>
                Want to deploy yourself?
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
