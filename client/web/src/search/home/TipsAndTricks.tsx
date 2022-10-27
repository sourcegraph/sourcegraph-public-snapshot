import React, { useCallback } from 'react'

import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Text, Card, Link } from '@sourcegraph/wildcard'

import { SearchExample } from './LoggedOutHomepage.constants'

import styles from './TipsAndTricks.module.scss'

interface TipsAndTricksProps extends TelemetryProps {
    examples: SearchExample[]
    moreLink: {
        href: string
        label: string
        trackEventName: string
    }
}

export const TipsAndTricks: React.FunctionComponent<React.PropsWithChildren<TipsAndTricksProps>> = ({
    moreLink,
    telemetryService,
    examples,
}) => {
    const searchExampleClicked = useCallback(
        (trackEventName: string) => (): void => telemetryService.log(trackEventName),
        [telemetryService]
    )
    return (
        <div className={classNames(styles.tipsAndTricks)}>
            <div className="d-flex align-items-center mb-2">
                <Text className={classNames('mr-2 pr-2', styles.tipsAndTricksTitle)}>Code Search Basics</Text>
            </div>
            <div className={styles.tipsAndTricksExamples}>
                {examples.map(example => (
                    <div key={example.query} className={styles.tipsAndTricksExample}>
                        {example.label}
                        <Card
                            as={Link}
                            to={example.to}
                            className={styles.tipsAndTricksCard}
                            onClick={searchExampleClicked(example.trackEventName)}
                        >
                            <SyntaxHighlightedSearchQuery query={example.query} />
                        </Card>
                    </div>
                ))}
            </div>
            <Link
                className={styles.tipsAndTricksMore}
                onClick={searchExampleClicked(moreLink.trackEventName)}
                to={moreLink.href}
            >
                {moreLink.label}
            </Link>
        </div>
    )
}
