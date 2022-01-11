import classNames from 'classnames'
import React from 'react'

import { ProductStatusBadge } from '@sourcegraph/wildcard'
import type { BaseProductStatusBadgeProps } from '@sourcegraph/wildcard/src/components/Badge'

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
export const FeedbackBadge: React.FunctionComponent<FeedbackBadgeProps> = ({
    className,
    status,
    tooltip,
    feedback: { mailto, text },
}) => (
    <div className={classNames(styles.feedbackBadge, className)}>
        <ProductStatusBadge tooltip={tooltip} status={status} className={styles.productStatusBadge} />
        <a href={`mailto:${mailto}`} className={styles.anchor} target="_blank" rel="noopener noreferrer">
            {text || 'Share feedback'}
        </a>
    </div>
)
