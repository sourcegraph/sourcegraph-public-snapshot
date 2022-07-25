import { FC } from 'react'

import classNames from 'classnames'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { Link } from '@sourcegraph/wildcard'

import styles from './InsightDecorationPopover.module.scss'

interface TokenInsight {
    id: string
    name: string
    url: string
}

interface Token {
    name: string
    insights: TokenInsight[]
}

interface InsightDecorationPopoverProps {
    tokens: Token[]
}

export const InsightDecorationPopover: FC<InsightDecorationPopoverProps> = ({ tokens }) => (
    <div className={styles.insightDecorationPopover}>
        {tokens.map(token => (
            <div key={token.name} className={styles.insightDecorationSection}>
                <div>
                    <span>{'{}'}</span>
                    <small className="ml-2">
                        <strong>{token.name}</strong>
                    </small>
                </div>
                <div className={classNames(styles.insightDecorationRow, styles.insightDecorationLineRef)}>
                    Insights referencing this line ({token.insights.length})
                </div>
                {token.insights.map(insight => (
                    <div key={insight.id} className={classNames(styles.insightDecorationRow)}>
                        <Link to={insight.url} className={styles.insightDecorationLink}>
                            {insight.name} <OpenInNewIcon size={12} />
                        </Link>
                    </div>
                ))}
            </div>
        ))}
    </div>
)
