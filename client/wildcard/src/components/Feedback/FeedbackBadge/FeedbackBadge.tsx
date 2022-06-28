import React from 'react'

import classNames from 'classnames'

import { ProductStatusBadge, Link, BaseProductStatusBadgeProps } from '@sourcegraph/wildcard'

import styles from './FeedbackBadge.module.scss'

interface FeedbackBadgeProps extends BaseProductStatusBadgeProps {
    /** Render a mailto href to share feedback */
    feedback: {
        mailto: string
        /** Defaults to 'Share feedback' */
        text?: string
    }
    className?: string
}

/**
 * An abstract UI component which renders a badge with specific status for feedback.
 */
export const FeedbackBadge: React.FunctionComponent<React.PropsWithChildren<FeedbackBadgeProps>> = ({
    className,
    status,
    tooltip,
    feedback: { mailto, text },
}) => (
    <div className={classNames(styles.feedbackBadge, className)}>
        <ProductStatusBadge tooltip={tooltip} status={status} className={styles.productStatusBadge} />
        <Link to={`mailto:${mailto}`} className={styles.anchor} target="_blank" rel="noopener noreferrer">
            {text || 'Share feedback'}
        </Link>
    </div>
)
