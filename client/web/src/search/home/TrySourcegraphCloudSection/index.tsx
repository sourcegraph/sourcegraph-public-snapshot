import { FunctionComponent, useEffect, useState } from 'react'

import { ButtonLink, H2, H3, Text, Link } from '@sourcegraph/wildcard'

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
        <div className={styles.wrapper}>
            <Illustration className={styles.illustration} />

            <div className={styles.content}>
                <div>
                    <H2 className={styles.tryCloud}>Try Sourcegraph Cloud.</H2>
                    <H3 className={styles.byline}>{byline}</H3>
                    <Text className={styles.signUp}>Sign up for a 30-day trial for your team.</Text>
                </div>

                <div className={styles.links}>
                    <ButtonLink
                        variant="secondary"
                        to="https://signup.sourcegraph.com"
                        className={styles.trialButton}
                        onClick={() => eventLogger.log('ClickedOnCloudCTA')}
                    >
                        Get free trial now
                    </ButtonLink>
                    <Link to="https://docs.sourcegraph.com" onClick={() => eventLogger.log('ClickedOnSelfHostedCTA')}>
                        <Text size="small">or try our self-hosted solution.</Text>
                    </Link>
                </div>
            </div>
        </div>
    )
}
