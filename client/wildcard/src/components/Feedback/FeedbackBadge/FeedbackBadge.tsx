import React from 'react'

import classNames from 'classnames'

import { ProductStatusBadge, type BaseProductStatusBadgeProps } from '../../Badge'
import { Link } from '../../Link'

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
    feedback: { mailto, text },
}) => (
    <div className={classNames(styles.feedbackBadge, className)}>
        <ProductStatusBadge status={status} className={styles.productStatusBadge} />
        <Link to={`mailto:${mailto}`} className={styles.anchor} target="_blank" rel="noopener noreferrer">
            {text || 'Share feedback'}
        </Link>
    </div>
)
