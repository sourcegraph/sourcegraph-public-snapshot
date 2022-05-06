import React, { useCallback } from 'react'

import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Card, Link } from '@sourcegraph/wildcard'

import { SearchExample } from './LoggedOutHomepage.constants'

import styles from './TipsAndTricks.module.scss'

interface TipsAndTricksProps extends TelemetryProps {
    title: string
    examples: SearchExample[]
    moreLink: {
        href: string
        label: string
        trackEventName: string
    }
}

export const TipsAndTricks: React.FunctionComponent<React.PropsWithChildren<TipsAndTricksProps>> = ({
    title,
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
            <div className={classNames('mb-2', styles.tipsAndTricksTitle)}>{title}</div>
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
