import { FunctionComponent, useEffect, useState } from 'react'

import { ButtonLink, H2, H3, Text, Link } from '@sourcegraph/wildcard'

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
                    >
                        Get free trial now
                    </ButtonLink>
                    <Link to="https://docs.sourcegraph.com">
                        <Text size="small">or try our self-hosted solution.</Text>
                    </Link>
                </div>
            </div>
        </div>
    )
}
